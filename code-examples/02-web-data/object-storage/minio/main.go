// minio-go 完整示例 — 对象存储 Bucket 管理 / 上传下载 / 预签名 URL / 分片上传 / Bucket 策略
// 演示：MinIO 对象存储的核心操作，包含内存模拟和真实 minio-go SDK 调用
// Go 1.22+ | 验证日期 2025-01-01
//
// Part A：纯内存模拟，直接运行理解对象存储核心概念
// Part B：连接真实 MinIO，需传入参数 'real'
//
// 运行方式：
//   go run ./minio/              # Part A：内存模拟
//   go run ./minio/ real         # Part B：连接真实 MinIO
//
// Part B 前置条件：
//   docker compose -f docker/docker-compose.yml up -d minio
//   API 地址：localhost:9000，控制台：localhost:9001
//   用户名：minioadmin，密码：minioadmin

package main

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"math/rand"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// ============================================================
// Part A：纯内存模拟对象存储核心概念
// ============================================================

func partA() {
	fmt.Println("=" + strings.Repeat("=", 59))
	fmt.Println("Part A：对象存储核心概念（内存模拟）")
	fmt.Println("=" + strings.Repeat("=", 59))

	store := NewMemoryObjectStore()

	// 1. Bucket 管理
	demoBucketManagement(store)

	// 2. 对象上传与下载
	demoObjectOperations(store)

	// 3. 预签名 URL 生成
	demoPresignedURL(store)

	// 4. 分片上传状态机
	demoMultipartUpload(store)

	// 5. Bucket 策略管理
	demoBucketPolicy(store)
}

// ============================================================
// 内存对象存储实现
// ============================================================

// ObjectMeta 对象元数据
type ObjectMeta struct {
	Key          string
	Size         int64
	ContentType  string
	ETag         string
	LastModified time.Time
	UserMeta     map[string]string
}

// StoredObject 存储的对象（数据 + 元数据）
type StoredObject struct {
	Meta ObjectMeta
	Data []byte
}

// BucketInfo Bucket 信息
type BucketInfo struct {
	Name      string
	CreatedAt time.Time
	Policy    string // 访问策略 JSON
}

// MultipartUpload 分片上传会话
type MultipartUpload struct {
	UploadID  string
	Bucket    string
	Key       string
	Parts     map[int]*UploadPart // partNumber -> part
	CreatedAt time.Time
	Completed bool
}

// UploadPart 上传的分片
type UploadPart struct {
	PartNumber int
	Data       []byte
	ETag       string
	Size       int64
}

// MemoryObjectStore 内存对象存储引擎
type MemoryObjectStore struct {
	mu       sync.RWMutex
	buckets  map[string]*BucketInfo
	objects  map[string]map[string]*StoredObject // bucket -> key -> object
	uploads  map[string]*MultipartUpload         // uploadID -> upload
	secretKey string
}

// NewMemoryObjectStore 创建内存对象存储
func NewMemoryObjectStore() *MemoryObjectStore {
	return &MemoryObjectStore{
		buckets:   make(map[string]*BucketInfo),
		objects:   make(map[string]map[string]*StoredObject),
		uploads:   make(map[string]*MultipartUpload),
		secretKey: "memory-store-secret-key-2025",
	}
}

// MakeBucket 创建 Bucket
func (s *MemoryObjectStore) MakeBucket(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.buckets[name]; exists {
		return fmt.Errorf("bucket %q already exists", name)
	}
	s.buckets[name] = &BucketInfo{Name: name, CreatedAt: time.Now()}
	s.objects[name] = make(map[string]*StoredObject)
	return nil
}

// ListBuckets 列出所有 Bucket
func (s *MemoryObjectStore) ListBuckets() []BucketInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]BucketInfo, 0, len(s.buckets))
	for _, b := range s.buckets {
		result = append(result, *b)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}

// BucketExists 检查 Bucket 是否存在
func (s *MemoryObjectStore) BucketExists(name string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, exists := s.buckets[name]
	return exists
}

// RemoveBucket 删除空 Bucket
func (s *MemoryObjectStore) RemoveBucket(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.buckets[name]; !exists {
		return fmt.Errorf("bucket %q not found", name)
	}
	if len(s.objects[name]) > 0 {
		return fmt.Errorf("bucket %q is not empty", name)
	}
	delete(s.buckets, name)
	delete(s.objects, name)
	return nil
}

// computeETag 计算对象的 ETag（基于 SHA256）
func computeETag(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:16])
}

// PutObject 上传对象
func (s *MemoryObjectStore) PutObject(bucket, key string, data []byte, contentType string, userMeta map[string]string) (*ObjectMeta, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.buckets[bucket]; !exists {
		return nil, fmt.Errorf("bucket %q not found", bucket)
	}
	meta := ObjectMeta{
		Key:          key,
		Size:         int64(len(data)),
		ContentType:  contentType,
		ETag:         computeETag(data),
		LastModified: time.Now(),
		UserMeta:     userMeta,
	}
	s.objects[bucket][key] = &StoredObject{Meta: meta, Data: append([]byte(nil), data...)}
	return &meta, nil
}

// GetObject 下载对象
func (s *MemoryObjectStore) GetObject(bucket, key string) (*StoredObject, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	objs, exists := s.objects[bucket]
	if !exists {
		return nil, fmt.Errorf("bucket %q not found", bucket)
	}
	obj, exists := objs[key]
	if !exists {
		return nil, fmt.Errorf("object %q not found in bucket %q", key, bucket)
	}
	return obj, nil
}

// StatObject 获取对象元数据
func (s *MemoryObjectStore) StatObject(bucket, key string) (*ObjectMeta, error) {
	obj, err := s.GetObject(bucket, key)
	if err != nil {
		return nil, err
	}
	return &obj.Meta, nil
}

// RemoveObject 删除对象
func (s *MemoryObjectStore) RemoveObject(bucket, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	objs, exists := s.objects[bucket]
	if !exists {
		return fmt.Errorf("bucket %q not found", bucket)
	}
	if _, exists := objs[key]; !exists {
		return fmt.Errorf("object %q not found", key)
	}
	delete(objs, key)
	return nil
}

// ListObjects 列出 Bucket 中的对象
func (s *MemoryObjectStore) ListObjects(bucket, prefix string) []ObjectMeta {
	s.mu.RLock()
	defer s.mu.RUnlock()
	objs, exists := s.objects[bucket]
	if !exists {
		return nil
	}
	result := make([]ObjectMeta, 0)
	for key, obj := range objs {
		if prefix == "" || strings.HasPrefix(key, prefix) {
			result = append(result, obj.Meta)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Key < result[j].Key
	})
	return result
}

// GeneratePresignedURL 生成预签名 URL
func (s *MemoryObjectStore) GeneratePresignedURL(bucket, key, method string, expiry time.Duration) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if _, exists := s.buckets[bucket]; !exists {
		return "", fmt.Errorf("bucket %q not found", bucket)
	}
	// 构造签名参数
	expiresAt := time.Now().Add(expiry).Unix()
	signData := fmt.Sprintf("%s\n%s\n%s\n%d", method, bucket, key, expiresAt)
	mac := hmac.New(sha256.New, []byte(s.secretKey))
	mac.Write([]byte(signData))
	signature := hex.EncodeToString(mac.Sum(nil))
	url := fmt.Sprintf("http://localhost:9000/%s/%s?X-Amz-Expires=%d&X-Amz-Signature=%s",
		bucket, key, int(expiry.Seconds()), signature[:32])
	return url, nil
}

// InitMultipartUpload 初始化分片上传
func (s *MemoryObjectStore) InitMultipartUpload(bucket, key string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.buckets[bucket]; !exists {
		return "", fmt.Errorf("bucket %q not found", bucket)
	}
	uploadID := fmt.Sprintf("upload-%d-%d", time.Now().UnixNano(), rand.Intn(10000))
	s.uploads[uploadID] = &MultipartUpload{
		UploadID:  uploadID,
		Bucket:    bucket,
		Key:       key,
		Parts:     make(map[int]*UploadPart),
		CreatedAt: time.Now(),
	}
	return uploadID, nil
}

// UploadParts 上传分片
func (s *MemoryObjectStore) UploadParts(uploadID string, partNumber int, data []byte) (*UploadPart, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	upload, exists := s.uploads[uploadID]
	if !exists {
		return nil, fmt.Errorf("upload %q not found", uploadID)
	}
	if upload.Completed {
		return nil, fmt.Errorf("upload %q already completed", uploadID)
	}
	part := &UploadPart{
		PartNumber: partNumber,
		Data:       append([]byte(nil), data...),
		ETag:       computeETag(data),
		Size:       int64(len(data)),
	}
	upload.Parts[partNumber] = part
	return part, nil
}

// CompleteMultipartUpload 完成分片上传，合并所有分片
func (s *MemoryObjectStore) CompleteMultipartUpload(uploadID string) (*ObjectMeta, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	upload, exists := s.uploads[uploadID]
	if !exists {
		return nil, fmt.Errorf("upload %q not found", uploadID)
	}
	// 按分片编号排序并合并数据
	partNumbers := make([]int, 0, len(upload.Parts))
	for pn := range upload.Parts {
		partNumbers = append(partNumbers, pn)
	}
	sort.Ints(partNumbers)
	var merged []byte
	for _, pn := range partNumbers {
		merged = append(merged, upload.Parts[pn].Data...)
	}
	// 存储合并后的对象
	meta := ObjectMeta{
		Key:          upload.Key,
		Size:         int64(len(merged)),
		ContentType:  "application/octet-stream",
		ETag:         computeETag(merged),
		LastModified: time.Now(),
	}
	s.objects[upload.Bucket][upload.Key] = &StoredObject{Meta: meta, Data: merged}
	upload.Completed = true
	delete(s.uploads, uploadID)
	return &meta, nil
}

// SetBucketPolicy 设置 Bucket 策略
func (s *MemoryObjectStore) SetBucketPolicy(bucket, policy string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	b, exists := s.buckets[bucket]
	if !exists {
		return fmt.Errorf("bucket %q not found", bucket)
	}
	b.Policy = policy
	return nil
}

// GetBucketPolicy 获取 Bucket 策略
func (s *MemoryObjectStore) GetBucketPolicy(bucket string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	b, exists := s.buckets[bucket]
	if !exists {
		return "", fmt.Errorf("bucket %q not found", bucket)
	}
	return b.Policy, nil
}

// ============================================================
// Part A 演示函数
// ============================================================

// demoBucketManagement 演示 Bucket 管理操作
func demoBucketManagement(store *MemoryObjectStore) {
	fmt.Println("\n--- 1. Bucket 管理 ---")

	// 创建 Bucket
	buckets := []string{"images", "documents", "backups"}
	for _, name := range buckets {
		if err := store.MakeBucket(name); err != nil {
			fmt.Printf("  创建 Bucket 失败: %v\n", err)
			continue
		}
		fmt.Printf("  创建 Bucket: %s ✅\n", name)
	}

	// 列出所有 Bucket
	fmt.Println("\n  所有 Bucket:")
	for _, b := range store.ListBuckets() {
		fmt.Printf("    - %s (创建时间: %s)\n", b.Name, b.CreatedAt.Format("2006-01-02 15:04:05"))
	}

	// 检查 Bucket 是否存在
	fmt.Printf("\n  Bucket 'images' 存在: %v\n", store.BucketExists("images"))
	fmt.Printf("  Bucket 'videos' 存在: %v\n", store.BucketExists("videos"))

	// 尝试删除非空 Bucket（后面会放入对象）
	fmt.Printf("\n  删除空 Bucket 'backups': ")
	if err := store.RemoveBucket("backups"); err != nil {
		fmt.Printf("失败 - %v\n", err)
	} else {
		fmt.Println("成功 ✅")
	}
}

// demoObjectOperations 演示对象上传与下载
func demoObjectOperations(store *MemoryObjectStore) {
	fmt.Println("\n--- 2. 对象上传与下载 ---")

	// 上传文本文件
	textContent := []byte("# Go 知识库\n\n这是一个 Go 语言学习知识库的示例文档。\n包含从入门到精通的完整内容。")
	meta, err := store.PutObject("documents", "guides/go-intro.md", textContent, "text/markdown",
		map[string]string{"author": "guide-go", "version": "1.0"})
	if err != nil {
		fmt.Printf("  上传失败: %v\n", err)
		return
	}
	fmt.Printf("  上传文件: %s (大小: %d bytes, ETag: %s)\n", meta.Key, meta.Size, meta.ETag)

	// 上传模拟图片（二进制数据）
	imageData := generateRandomBytes(2048)
	meta, err = store.PutObject("images", "avatars/user-001.png", imageData, "image/png", nil)
	if err != nil {
		fmt.Printf("  上传失败: %v\n", err)
		return
	}
	fmt.Printf("  上传图片: %s (大小: %d bytes, ETag: %s)\n", meta.Key, meta.Size, meta.ETag)

	// 上传更多文件用于列表演示
	files := []struct {
		key         string
		contentType string
		size        int
	}{
		{"avatars/user-002.png", "image/png", 1536},
		{"banners/home.jpg", "image/jpeg", 4096},
		{"banners/about.jpg", "image/jpeg", 3072},
	}
	for _, f := range files {
		data := generateRandomBytes(f.size)
		store.PutObject("images", f.key, data, f.contentType, nil)
	}

	// 下载文件并验证内容
	fmt.Println("\n  下载并验证文件:")
	obj, err := store.GetObject("documents", "guides/go-intro.md")
	if err != nil {
		fmt.Printf("  下载失败: %v\n", err)
		return
	}
	fmt.Printf("    文件: %s\n", obj.Meta.Key)
	fmt.Printf("    大小: %d bytes\n", obj.Meta.Size)
	fmt.Printf("    类型: %s\n", obj.Meta.ContentType)
	fmt.Printf("    内容预览: %s...\n", string(obj.Data[:40]))
	contentMatch := bytes.Equal(obj.Data, textContent)
	fmt.Printf("    内容校验: %v ✅\n", contentMatch)

	// 获取对象元数据（不下载数据）
	fmt.Println("\n  获取对象元数据（StatObject）:")
	stat, _ := store.StatObject("images", "avatars/user-001.png")
	fmt.Printf("    Key: %s\n", stat.Key)
	fmt.Printf("    Size: %d bytes\n", stat.Size)
	fmt.Printf("    ContentType: %s\n", stat.ContentType)
	fmt.Printf("    LastModified: %s\n", stat.LastModified.Format("2006-01-02 15:04:05"))

	// 列出对象（带前缀过滤）
	fmt.Println("\n  列出 images Bucket 中的所有对象:")
	for _, m := range store.ListObjects("images", "") {
		fmt.Printf("    %s (%d bytes, %s)\n", m.Key, m.Size, m.ContentType)
	}

	fmt.Println("\n  列出 images/avatars/ 前缀的对象:")
	for _, m := range store.ListObjects("images", "avatars/") {
		fmt.Printf("    %s (%d bytes)\n", m.Key, m.Size)
	}

	// 删除对象
	fmt.Println("\n  删除对象:")
	if err := store.RemoveObject("images", "banners/about.jpg"); err != nil {
		fmt.Printf("    删除失败: %v\n", err)
	} else {
		fmt.Println("    删除 banners/about.jpg ✅")
	}
	fmt.Printf("  删除后 images Bucket 对象数: %d\n", len(store.ListObjects("images", "")))
}

// demoPresignedURL 演示预签名 URL 生成
func demoPresignedURL(store *MemoryObjectStore) {
	fmt.Println("\n--- 3. 预签名 URL ---")

	// 生成下载预签名 URL
	getURL, err := store.GeneratePresignedURL("images", "avatars/user-001.png", "GET", 15*time.Minute)
	if err != nil {
		fmt.Printf("  生成失败: %v\n", err)
		return
	}
	fmt.Printf("  下载预签名 URL (有效期 15 分钟):\n    %s\n", getURL)

	// 生成上传预签名 URL
	putURL, err := store.GeneratePresignedURL("images", "uploads/new-file.png", "PUT", 5*time.Minute)
	if err != nil {
		fmt.Printf("  生成失败: %v\n", err)
		return
	}
	fmt.Printf("\n  上传预签名 URL (有效期 5 分钟):\n    %s\n", putURL)

	fmt.Println("\n  预签名 URL 工作流程:")
	fmt.Println("    1. 客户端请求后端生成预签名 URL")
	fmt.Println("    2. 后端使用 Secret Key 计算 HMAC-SHA256 签名")
	fmt.Println("    3. 客户端使用预签名 URL 直接上传/下载文件到 MinIO")
	fmt.Println("    4. MinIO 验证签名和过期时间，处理请求")
	fmt.Println("    ⚠️  文件不经过后端服务器，减少带宽压力")
}

// demoMultipartUpload 演示分片上传状态机
func demoMultipartUpload(store *MemoryObjectStore) {
	fmt.Println("\n--- 4. 分片上传 ---")

	// 模拟一个 15MB 的大文件
	totalSize := 15 * 1024 * 1024
	partSize := 5 * 1024 * 1024
	fileData := generateRandomBytes(totalSize)

	fmt.Printf("  文件总大小: %d MB\n", totalSize/(1024*1024))
	fmt.Printf("  分片大小: %d MB\n", partSize/(1024*1024))
	fmt.Printf("  分片数量: %d\n", (totalSize+partSize-1)/partSize)

	// 1. 初始化分片上传
	uploadID, err := store.InitMultipartUpload("documents", "large-files/dataset.bin")
	if err != nil {
		fmt.Printf("  初始化失败: %v\n", err)
		return
	}
	fmt.Printf("\n  初始化分片上传: UploadID=%s\n", uploadID)

	// 2. 上传各分片
	fmt.Println("  上传分片:")
	partNum := 1
	for offset := 0; offset < totalSize; offset += partSize {
		end := offset + partSize
		if end > totalSize {
			end = totalSize
		}
		chunk := fileData[offset:end]
		part, err := store.UploadParts(uploadID, partNum, chunk)
		if err != nil {
			fmt.Printf("    分片 %d 上传失败: %v\n", partNum, err)
			return
		}
		fmt.Printf("    分片 %d: %d bytes, ETag=%s ✅\n", part.PartNumber, part.Size, part.ETag)
		partNum++
	}

	// 3. 完成分片上传（合并）
	meta, err := store.CompleteMultipartUpload(uploadID)
	if err != nil {
		fmt.Printf("  合并失败: %v\n", err)
		return
	}
	fmt.Printf("\n  分片合并完成: %s (%d bytes, ETag=%s) ✅\n", meta.Key, meta.Size, meta.ETag)

	// 4. 验证合并后的文件
	obj, _ := store.GetObject("documents", "large-files/dataset.bin")
	dataMatch := bytes.Equal(obj.Data, fileData)
	fmt.Printf("  数据完整性校验: %v ✅\n", dataMatch)
}

// demoBucketPolicy 演示 Bucket 策略管理
func demoBucketPolicy(store *MemoryObjectStore) {
	fmt.Println("\n--- 5. Bucket 策略 ---")

	// 设置公开只读策略
	publicReadPolicy := `{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Principal": {"AWS": ["*"]},
    "Action": ["s3:GetObject"],
    "Resource": ["arn:aws:s3:::images/*"]
  }]
}`
	if err := store.SetBucketPolicy("images", publicReadPolicy); err != nil {
		fmt.Printf("  设置策略失败: %v\n", err)
		return
	}
	fmt.Println("  设置 images Bucket 为公开只读 ✅")

	// 读取策略
	policy, _ := store.GetBucketPolicy("images")
	fmt.Printf("  当前策略:\n%s\n", policy)

	fmt.Println("\n  常见 Bucket 策略类型:")
	fmt.Println("    - 公开只读: 允许任何人下载（静态资源托管）")
	fmt.Println("    - 公开读写: 允许任何人上传和下载（不推荐）")
	fmt.Println("    - 私有: 仅通过预签名 URL 或认证访问（默认）")
	fmt.Println("    - 条件策略: 限制 IP 范围、Referer 等")
}

// generateRandomBytes 生成指定大小的随机字节数据
func generateRandomBytes(size int) []byte {
	data := make([]byte, size)
	for i := range data {
		data[i] = byte(rand.Intn(256))
	}
	return data
}

// ============================================================
// Part B：连接真实 MinIO（需要 Docker）
// ============================================================

func partB() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("Part B：连接真实 MinIO（minio-go）")
	fmt.Println(strings.Repeat("=", 60))

	ctx := context.Background()

	// 创建 MinIO 客户端
	client, err := minio.New("localhost:9000", &minio.Options{
		Creds:  credentials.NewStaticV4("minioadmin", "minioadmin", ""),
		Secure: false,
	})
	if err != nil {
		fmt.Printf("❌ 创建 MinIO 客户端失败: %v\n", err)
		return
	}

	// 测试连接（通过列出 Bucket 验证）
	_, err = client.ListBuckets(ctx)
	if err != nil {
		fmt.Printf("❌ 无法连接 MinIO: %v\n", err)
		fmt.Println("请先启动 MinIO: docker compose -f docker/docker-compose.yml up -d minio")
		return
	}
	fmt.Println("✅ MinIO 连接成功")

	// 1. Bucket 操作
	demoRealBucketOps(ctx, client)

	// 2. 对象上传与下载
	demoRealObjectOps(ctx, client)

	// 3. 预签名 URL
	demoRealPresignedURL(ctx, client)

	// 4. 对象列表与前缀过滤
	demoRealListObjects(ctx, client)

	// 清理测试数据
	cleanupMinioTestData(ctx, client)
}

// demoRealBucketOps 演示真实 Bucket 操作
func demoRealBucketOps(ctx context.Context, client *minio.Client) {
	fmt.Println("\n--- 1. Bucket 操作（minio-go） ---")

	bucketName := "demo-guide-go"

	// 创建 Bucket
	exists, err := client.BucketExists(ctx, bucketName)
	if err != nil {
		fmt.Printf("  检查 Bucket 失败: %v\n", err)
		return
	}
	if !exists {
		err = client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			fmt.Printf("  创建 Bucket 失败: %v\n", err)
			return
		}
		fmt.Printf("  创建 Bucket: %s ✅\n", bucketName)
	} else {
		fmt.Printf("  Bucket %s 已存在 ✅\n", bucketName)
	}

	// 列出所有 Bucket
	buckets, _ := client.ListBuckets(ctx)
	fmt.Println("  所有 Bucket:")
	for _, b := range buckets {
		fmt.Printf("    - %s (创建时间: %s)\n", b.Name, b.CreationDate.Format("2006-01-02 15:04:05"))
	}
}

// demoRealObjectOps 演示真实对象上传与下载
func demoRealObjectOps(ctx context.Context, client *minio.Client) {
	fmt.Println("\n--- 2. 对象上传与下载（minio-go） ---")

	bucketName := "demo-guide-go"

	// 上传文本文件
	content := []byte("# Go 知识库示例文件\n\n这是通过 minio-go SDK 上传的测试文件。\n验证日期: 2025-01-01")
	reader := bytes.NewReader(content)
	info, err := client.PutObject(ctx, bucketName, "docs/readme.md", reader, int64(len(content)),
		minio.PutObjectOptions{
			ContentType: "text/markdown",
			UserMetadata: map[string]string{
				"Author":  "guide-go",
				"Version": "1.0",
			},
		})
	if err != nil {
		fmt.Printf("  上传失败: %v\n", err)
		return
	}
	fmt.Printf("  上传文件: %s (大小: %d bytes, ETag: %s)\n", info.Key, info.Size, info.ETag)

	// 上传二进制数据（模拟图片）
	imageData := generateRandomBytes(4096)
	imgReader := bytes.NewReader(imageData)
	info, err = client.PutObject(ctx, bucketName, "images/test-avatar.png", imgReader, int64(len(imageData)),
		minio.PutObjectOptions{ContentType: "image/png"})
	if err != nil {
		fmt.Printf("  上传图片失败: %v\n", err)
		return
	}
	fmt.Printf("  上传图片: %s (大小: %d bytes)\n", info.Key, info.Size)

	// 下载文件并验证
	obj, err := client.GetObject(ctx, bucketName, "docs/readme.md", minio.GetObjectOptions{})
	if err != nil {
		fmt.Printf("  下载失败: %v\n", err)
		return
	}
	defer obj.Close()

	downloaded, err := io.ReadAll(obj)
	if err != nil {
		fmt.Printf("  读取失败: %v\n", err)
		return
	}
	fmt.Printf("  下载文件: docs/readme.md (%d bytes)\n", len(downloaded))
	fmt.Printf("  内容校验: %v ✅\n", bytes.Equal(downloaded, content))

	// 获取对象元数据
	stat, err := client.StatObject(ctx, bucketName, "docs/readme.md", minio.StatObjectOptions{})
	if err != nil {
		fmt.Printf("  获取元数据失败: %v\n", err)
		return
	}
	fmt.Printf("  元数据: ContentType=%s, Size=%d, LastModified=%s\n",
		stat.ContentType, stat.Size, stat.LastModified.Format("2006-01-02 15:04:05"))
	if author, ok := stat.UserMetadata["Author"]; ok {
		fmt.Printf("  自定义元数据: Author=%s\n", author)
	}
}

// demoRealPresignedURL 演示真实预签名 URL
func demoRealPresignedURL(ctx context.Context, client *minio.Client) {
	fmt.Println("\n--- 3. 预签名 URL（minio-go） ---")

	bucketName := "demo-guide-go"

	// 生成下载预签名 URL（有效期 15 分钟）
	getURL, err := client.PresignedGetObject(ctx, bucketName, "docs/readme.md", 15*time.Minute, nil)
	if err != nil {
		fmt.Printf("  生成下载 URL 失败: %v\n", err)
		return
	}
	fmt.Printf("  下载预签名 URL (15min):\n    %s\n", getURL.String())

	// 生成上传预签名 URL（有效期 5 分钟）
	putURL, err := client.PresignedPutObject(ctx, bucketName, "uploads/new-file.txt", 5*time.Minute)
	if err != nil {
		fmt.Printf("  生成上传 URL 失败: %v\n", err)
		return
	}
	fmt.Printf("\n  上传预签名 URL (5min):\n    %s\n", putURL.String())
}

// demoRealListObjects 演示对象列表与前缀过滤
func demoRealListObjects(ctx context.Context, client *minio.Client) {
	fmt.Println("\n--- 4. 对象列表与前缀过滤 ---")

	bucketName := "demo-guide-go"

	// 上传更多测试文件
	testFiles := []struct {
		key         string
		contentType string
	}{
		{"images/banner-home.jpg", "image/jpeg"},
		{"images/banner-about.jpg", "image/jpeg"},
		{"docs/guide-advanced.md", "text/markdown"},
	}
	for _, f := range testFiles {
		data := generateRandomBytes(1024)
		client.PutObject(ctx, bucketName, f.key, bytes.NewReader(data), int64(len(data)),
			minio.PutObjectOptions{ContentType: f.contentType})
	}

	// 列出所有对象
	fmt.Println("  所有对象:")
	for obj := range client.ListObjects(ctx, bucketName, minio.ListObjectsOptions{Recursive: true}) {
		if obj.Err != nil {
			fmt.Printf("    错误: %v\n", obj.Err)
			continue
		}
		fmt.Printf("    %s (%d bytes, %s)\n", obj.Key, obj.Size, obj.ContentType)
	}

	// 按前缀过滤
	fmt.Println("\n  images/ 前缀的对象:")
	for obj := range client.ListObjects(ctx, bucketName, minio.ListObjectsOptions{Prefix: "images/", Recursive: true}) {
		if obj.Err != nil {
			continue
		}
		fmt.Printf("    %s (%d bytes)\n", obj.Key, obj.Size)
	}
}

// cleanupMinioTestData 清理测试数据
func cleanupMinioTestData(ctx context.Context, client *minio.Client) {
	bucketName := "demo-guide-go"

	// 删除所有对象
	for obj := range client.ListObjects(ctx, bucketName, minio.ListObjectsOptions{Recursive: true}) {
		if obj.Err != nil {
			continue
		}
		client.RemoveObject(ctx, bucketName, obj.Key, minio.RemoveObjectOptions{})
	}

	// 删除 Bucket
	client.RemoveBucket(ctx, bucketName)
	fmt.Println("\n  测试数据已清理 ✅")
}

// ============================================================
// main 入口
// ============================================================

func main() {
	// Part A：纯内存模拟，直接运行理解原理
	partA()

	// Part B：连接真实 MinIO，需传入参数 'real'
	if len(os.Args) > 1 && os.Args[1] == "real" {
		partB()
	}
}

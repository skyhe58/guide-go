// S3 对象存储 — AWS SDK for Go v2 完整示例
// 演示：Bucket 管理 / 文件上传下载 / 预签名 URL / 分片上传 / 删除操作
// Go 1.22+ | 验证日期 2025-01-01
//
// Part A：纯内存模拟 S3 兼容对象存储，直接运行理解核心概念
// Part B：连接 LocalStack S3，需传入参数 'real'
//
// 运行方式：
//   go run ./s3/              # Part A：内存模拟
//   go run ./s3/ real         # Part B：连接 LocalStack S3
//
// Part B 前置条件：
//   docker compose -f docker/docker-compose.localstack.yml up -d
//   端点地址：http://localhost:4566
//   凭证：test / test（LocalStack 接受任意凭证）

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

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// ============================================================
// Part A：纯内存模拟 S3 兼容对象存储
// ============================================================

// ObjectMeta 对象元数据（模拟 S3 对象头信息）
type ObjectMeta struct {
	Key          string
	Size         int64
	ContentType  string
	ETag         string // 内容哈希，用于完整性校验
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
}

// MultipartUpload 分片上传会话
type MultipartUpload struct {
	UploadID  string
	Bucket    string
	Key       string
	Parts     map[int]*UploadPart
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

// InMemoryS3 内存 S3 兼容对象存储引擎
// 模拟 AWS S3 的核心操作：Bucket 管理、对象 CRUD、预签名 URL、分片上传
type InMemoryS3 struct {
	mu        sync.RWMutex
	buckets   map[string]*BucketInfo
	objects   map[string]map[string]*StoredObject // bucket -> key -> object
	uploads   map[string]*MultipartUpload         // uploadID -> upload
	secretKey string                              // 用于预签名 URL 的密钥
}

// NewInMemoryS3 创建内存 S3 实例
func NewInMemoryS3() *InMemoryS3 {
	return &InMemoryS3{
		buckets:   make(map[string]*BucketInfo),
		objects:   make(map[string]map[string]*StoredObject),
		uploads:   make(map[string]*MultipartUpload),
		secretKey: "aws-s3-demo-secret-key-2025",
	}
}

// computeETag 计算对象 ETag（基于 SHA256 前 16 字节）
func computeETag(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:16])
}

// CreateBucket 创建 Bucket
func (s *InMemoryS3) CreateBucket(name string) error {
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
func (s *InMemoryS3) ListBuckets() []BucketInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]BucketInfo, 0, len(s.buckets))
	for _, b := range s.buckets {
		result = append(result, *b)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })
	return result
}

// DeleteBucket 删除空 Bucket
func (s *InMemoryS3) DeleteBucket(name string) error {
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

// PutObject 上传对象
func (s *InMemoryS3) PutObject(bucket, key string, data []byte, contentType string) (*ObjectMeta, error) {
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
	}
	s.objects[bucket][key] = &StoredObject{Meta: meta, Data: append([]byte(nil), data...)}
	return &meta, nil
}

// GetObject 下载对象
func (s *InMemoryS3) GetObject(bucket, key string) (*StoredObject, error) {
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

// DeleteObject 删除对象
func (s *InMemoryS3) DeleteObject(bucket, key string) error {
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

// ListObjects 列出 Bucket 中的对象（支持前缀过滤）
func (s *InMemoryS3) ListObjects(bucket, prefix string) []ObjectMeta {
	s.mu.RLock()
	defer s.mu.RUnlock()
	objs, exists := s.objects[bucket]
	if !exists {
		return nil
	}
	var result []ObjectMeta
	for key, obj := range objs {
		if prefix == "" || strings.HasPrefix(key, prefix) {
			result = append(result, obj.Meta)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Key < result[j].Key })
	return result
}

// GeneratePresignedURL 生成预签名 URL（模拟 SigV4 签名）
func (s *InMemoryS3) GeneratePresignedURL(bucket, key, method string, expiry time.Duration) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if _, exists := s.buckets[bucket]; !exists {
		return "", fmt.Errorf("bucket %q not found", bucket)
	}
	// 模拟 AWS SigV4 签名过程
	expiresAt := time.Now().Add(expiry).Unix()
	signData := fmt.Sprintf("%s\n%s\n%s\n%d", method, bucket, key, expiresAt)
	mac := hmac.New(sha256.New, []byte(s.secretKey))
	mac.Write([]byte(signData))
	signature := hex.EncodeToString(mac.Sum(nil))

	url := fmt.Sprintf("http://localhost:4566/%s/%s?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Expires=%d&X-Amz-Signature=%s",
		bucket, key, int(expiry.Seconds()), signature[:32])
	return url, nil
}

// InitMultipartUpload 初始化分片上传
func (s *InMemoryS3) InitMultipartUpload(bucket, key string) (string, error) {
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

// UploadPartData 上传分片数据
func (s *InMemoryS3) UploadPartData(uploadID string, partNumber int, data []byte) (*UploadPart, error) {
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
func (s *InMemoryS3) CompleteMultipartUpload(uploadID string) (*ObjectMeta, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	upload, exists := s.uploads[uploadID]
	if !exists {
		return nil, fmt.Errorf("upload %q not found", uploadID)
	}
	// 按分片编号排序并合并
	partNumbers := make([]int, 0, len(upload.Parts))
	for pn := range upload.Parts {
		partNumbers = append(partNumbers, pn)
	}
	sort.Ints(partNumbers)

	var merged []byte
	for _, pn := range partNumbers {
		merged = append(merged, upload.Parts[pn].Data...)
	}

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

// generateRandomBytes 生成随机字节数据
func generateRandomBytes(size int) []byte {
	data := make([]byte, size)
	for i := range data {
		data[i] = byte(rand.Intn(256))
	}
	return data
}

// ============================================================
// Part A 演示
// ============================================================

func partA() {
	fmt.Println("=" + strings.Repeat("=", 59))
	fmt.Println("Part A：S3 对象存储核心概念（内存模拟）")
	fmt.Println("=" + strings.Repeat("=", 59))

	store := NewInMemoryS3()

	// --- 1. Bucket 管理 ---
	fmt.Println("\n--- 1. Bucket 管理 ---")
	for _, name := range []string{"app-images", "app-documents", "app-backups"} {
		if err := store.CreateBucket(name); err != nil {
			fmt.Printf("  创建失败: %v\n", err)
			continue
		}
		fmt.Printf("  创建 Bucket: %s ✅\n", name)
	}

	fmt.Println("\n  所有 Bucket:")
	for _, b := range store.ListBuckets() {
		fmt.Printf("    - %s (创建: %s)\n", b.Name, b.CreatedAt.Format("15:04:05"))
	}

	// 删除空 Bucket
	if err := store.DeleteBucket("app-backups"); err == nil {
		fmt.Println("\n  删除空 Bucket app-backups ✅")
	}

	// --- 2. 对象上传与下载 ---
	fmt.Println("\n--- 2. 对象上传与下载 ---")
	content := []byte("# Go 知识库\n\n这是通过 S3 API 上传的示例文档。\n包含 AWS SDK for Go v2 的完整用法。")
	meta, _ := store.PutObject("app-documents", "guides/go-aws.md", content, "text/markdown")
	fmt.Printf("  上传: %s (%d bytes, ETag: %s)\n", meta.Key, meta.Size, meta.ETag)

	imgData := generateRandomBytes(4096)
	meta, _ = store.PutObject("app-images", "avatars/user-001.png", imgData, "image/png")
	fmt.Printf("  上传: %s (%d bytes, ETag: %s)\n", meta.Key, meta.Size, meta.ETag)

	// 更多文件
	store.PutObject("app-images", "avatars/user-002.png", generateRandomBytes(2048), "image/png")
	store.PutObject("app-images", "banners/home.jpg", generateRandomBytes(8192), "image/jpeg")

	// 下载并验证
	obj, _ := store.GetObject("app-documents", "guides/go-aws.md")
	fmt.Printf("\n  下载: %s (%d bytes)\n", obj.Meta.Key, obj.Meta.Size)
	fmt.Printf("  内容校验: %v ✅\n", bytes.Equal(obj.Data, content))

	// 列出对象
	fmt.Println("\n  app-images 所有对象:")
	for _, m := range store.ListObjects("app-images", "") {
		fmt.Printf("    %s (%d bytes, %s)\n", m.Key, m.Size, m.ContentType)
	}

	fmt.Println("\n  app-images/avatars/ 前缀:")
	for _, m := range store.ListObjects("app-images", "avatars/") {
		fmt.Printf("    %s (%d bytes)\n", m.Key, m.Size)
	}

	// 删除对象
	store.DeleteObject("app-images", "banners/home.jpg")
	fmt.Printf("\n  删除 banners/home.jpg 后剩余 %d 个对象\n", len(store.ListObjects("app-images", "")))

	// --- 3. 预签名 URL ---
	fmt.Println("\n--- 3. 预签名 URL ---")
	getURL, _ := store.GeneratePresignedURL("app-images", "avatars/user-001.png", "GET", 15*time.Minute)
	fmt.Printf("  下载 URL (15min):\n    %s\n", getURL)

	putURL, _ := store.GeneratePresignedURL("app-images", "uploads/new.png", "PUT", 5*time.Minute)
	fmt.Printf("  上传 URL (5min):\n    %s\n", putURL)

	fmt.Println("\n  预签名 URL 工作原理:")
	fmt.Println("    1. 后端使用 Secret Key 计算 HMAC-SHA256 签名")
	fmt.Println("    2. 签名参数: HTTP 方法 + Bucket + Key + 过期时间")
	fmt.Println("    3. 客户端使用 URL 直接与 S3 交互（文件直传）")
	fmt.Println("    4. S3 验证签名和过期时间")

	// --- 4. 分片上传 ---
	fmt.Println("\n--- 4. 分片上传 ---")
	totalSize := 15 * 1024 * 1024 // 15MB
	partSize := 5 * 1024 * 1024   // 5MB（S3 最小分片大小）
	fileData := generateRandomBytes(totalSize)

	fmt.Printf("  文件大小: %d MB, 分片大小: %d MB\n", totalSize/(1024*1024), partSize/(1024*1024))

	uploadID, _ := store.InitMultipartUpload("app-documents", "large/dataset.bin")
	fmt.Printf("  初始化分片上传: UploadID=%s\n", uploadID[:20]+"...")

	partNum := 1
	for offset := 0; offset < totalSize; offset += partSize {
		end := offset + partSize
		if end > totalSize {
			end = totalSize
		}
		part, _ := store.UploadPartData(uploadID, partNum, fileData[offset:end])
		fmt.Printf("    分片 %d: %d bytes, ETag=%s ✅\n", part.PartNumber, part.Size, part.ETag[:16]+"...")
		partNum++
	}

	completedMeta, _ := store.CompleteMultipartUpload(uploadID)
	fmt.Printf("  合并完成: %s (%d bytes) ✅\n", completedMeta.Key, completedMeta.Size)

	// 验证数据完整性
	merged, _ := store.GetObject("app-documents", "large/dataset.bin")
	fmt.Printf("  数据完整性校验: %v ✅\n", bytes.Equal(merged.Data, fileData))
}

// ============================================================
// Part B：连接 LocalStack S3（AWS SDK for Go v2）
// ============================================================

func partB() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("Part B：连接 LocalStack S3（AWS SDK v2）")
	fmt.Println(strings.Repeat("=", 60))

	ctx := context.Background()

	// 加载 AWS 配置（指向 LocalStack）
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion("us-east-1"),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider("test", "test", ""),
		),
	)
	if err != nil {
		fmt.Printf("❌ 加载 AWS 配置失败: %v\n", err)
		return
	}

	// 创建 S3 客户端（指向 LocalStack 端点）
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String("http://localhost:4566")
		o.UsePathStyle = true // LocalStack 需要路径风格
	})

	// 测试连接
	_, err = client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		fmt.Printf("❌ 连接 LocalStack 失败: %v\n", err)
		fmt.Println("请先启动: docker compose -f docker/docker-compose.localstack.yml up -d")
		return
	}
	fmt.Println("✅ LocalStack S3 连接成功")

	bucketName := "demo-guide-go-s3"

	// --- 1. 创建 Bucket ---
	demoRealCreateBucket(ctx, client, bucketName)

	// --- 2. 上传与下载 ---
	demoRealPutGetObject(ctx, client, bucketName)

	// --- 3. 预签名 URL ---
	demoRealPresignedURL(ctx, client, cfg, bucketName)

	// --- 4. 列出与删除 ---
	demoRealListDelete(ctx, client, bucketName)

	// --- 5. 清理 ---
	cleanupS3(ctx, client, bucketName)
}

// demoRealCreateBucket 创建 Bucket
func demoRealCreateBucket(ctx context.Context, client *s3.Client, bucket string) {
	fmt.Println("\n--- 1. 创建 Bucket ---")

	_, err := client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		fmt.Printf("  创建 Bucket 失败（可能已存在）: %v\n", err)
	} else {
		fmt.Printf("  创建 Bucket: %s ✅\n", bucket)
	}

	// 列出所有 Bucket
	output, _ := client.ListBuckets(ctx, &s3.ListBucketsInput{})
	fmt.Println("  所有 Bucket:")
	for _, b := range output.Buckets {
		fmt.Printf("    - %s (创建: %s)\n", *b.Name, b.CreationDate.Format("2006-01-02 15:04:05"))
	}
}

// demoRealPutGetObject 上传与下载对象
func demoRealPutGetObject(ctx context.Context, client *s3.Client, bucket string) {
	fmt.Println("\n--- 2. 上传与下载对象 ---")

	// 上传文本文件
	content := []byte("# AWS S3 示例\n\n通过 AWS SDK for Go v2 上传到 LocalStack S3。\n验证日期: 2025-01-01")
	_, err := client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String("docs/readme.md"),
		Body:        bytes.NewReader(content),
		ContentType: aws.String("text/markdown"),
	})
	if err != nil {
		fmt.Printf("  上传失败: %v\n", err)
		return
	}
	fmt.Printf("  上传: docs/readme.md (%d bytes) ✅\n", len(content))

	// 上传二进制数据
	imgData := generateRandomBytes(4096)
	_, err = client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String("images/avatar.png"),
		Body:        bytes.NewReader(imgData),
		ContentType: aws.String("image/png"),
	})
	if err != nil {
		fmt.Printf("  上传图片失败: %v\n", err)
		return
	}
	fmt.Printf("  上传: images/avatar.png (%d bytes) ✅\n", len(imgData))

	// 下载并验证
	getOutput, err := client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String("docs/readme.md"),
	})
	if err != nil {
		fmt.Printf("  下载失败: %v\n", err)
		return
	}
	defer getOutput.Body.Close()

	downloaded, _ := io.ReadAll(getOutput.Body)
	fmt.Printf("  下载: docs/readme.md (%d bytes)\n", len(downloaded))
	fmt.Printf("  内容校验: %v ✅\n", bytes.Equal(downloaded, content))
	fmt.Printf("  ContentType: %s\n", *getOutput.ContentType)
}

// demoRealPresignedURL 预签名 URL
func demoRealPresignedURL(ctx context.Context, client *s3.Client, cfg aws.Config, bucket string) {
	fmt.Println("\n--- 3. 预签名 URL ---")

	// 创建预签名客户端
	presignClient := s3.NewPresignClient(client)

	// 生成下载预签名 URL
	getPresign, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String("docs/readme.md"),
	}, s3.WithPresignExpires(15*time.Minute))
	if err != nil {
		fmt.Printf("  生成下载 URL 失败: %v\n", err)
		return
	}
	// 截断显示
	urlDisplay := getPresign.URL
	if len(urlDisplay) > 100 {
		urlDisplay = urlDisplay[:100] + "..."
	}
	fmt.Printf("  下载预签名 URL (15min):\n    %s\n", urlDisplay)

	// 生成上传预签名 URL
	putPresign, err := presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String("uploads/new-file.txt"),
	}, s3.WithPresignExpires(5*time.Minute))
	if err != nil {
		fmt.Printf("  生成上传 URL 失败: %v\n", err)
		return
	}
	urlDisplay = putPresign.URL
	if len(urlDisplay) > 100 {
		urlDisplay = urlDisplay[:100] + "..."
	}
	fmt.Printf("  上传预签名 URL (5min):\n    %s\n", urlDisplay)
}

// demoRealListDelete 列出与删除对象
func demoRealListDelete(ctx context.Context, client *s3.Client, bucket string) {
	fmt.Println("\n--- 4. 列出与删除对象 ---")

	// 上传更多测试文件
	for _, key := range []string{"images/banner-1.jpg", "images/banner-2.jpg", "docs/guide.md"} {
		data := generateRandomBytes(1024)
		client.PutObject(ctx, &s3.PutObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
			Body:   bytes.NewReader(data),
		})
	}

	// 列出所有对象
	listOutput, _ := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
	})
	fmt.Println("  所有对象:")
	for _, obj := range listOutput.Contents {
		fmt.Printf("    %s (%d bytes)\n", *obj.Key, *obj.Size)
	}

	// 按前缀过滤
	listOutput, _ = client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		Prefix: aws.String("images/"),
	})
	fmt.Printf("\n  images/ 前缀 (%d 个):\n", len(listOutput.Contents))
	for _, obj := range listOutput.Contents {
		fmt.Printf("    %s (%d bytes)\n", *obj.Key, *obj.Size)
	}

	// 删除单个对象
	_, err := client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String("images/banner-2.jpg"),
	})
	if err != nil {
		fmt.Printf("  删除失败: %v\n", err)
	} else {
		fmt.Println("\n  删除 images/banner-2.jpg ✅")
	}
}

// cleanupS3 清理测试数据
func cleanupS3(ctx context.Context, client *s3.Client, bucket string) {
	fmt.Println("\n--- 清理测试数据 ---")

	// 删除所有对象
	listOutput, _ := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
	})
	for _, obj := range listOutput.Contents {
		client.DeleteObject(ctx, &s3.DeleteObjectInput{
			Bucket: aws.String(bucket),
			Key:    obj.Key,
		})
	}

	// 删除 Bucket
	client.DeleteBucket(ctx, &s3.DeleteBucketInput{
		Bucket: aws.String(bucket),
	})
	fmt.Println("  测试数据已清理 ✅")
}

func main() {
	// Part A：纯内存模拟，直接运行理解原理
	partA()

	// Part B：连接 LocalStack S3，需传入参数 'real'
	if len(os.Args) > 1 && os.Args[1] == "real" {
		partB()
	}
}

/**
 * VitePress 代码链接环境动态切换插件
 *
 * - pnpm build: 保持 GitHub URL 原样输出（跳转到 GitHub 仓库源码）
 * - pnpm dev:   重写为编辑器协议 URL，点击直接在本地编辑器中打开代码
 *
 * 通过环境变量 EDITOR_SCHEME 控制使用的编辑器（默认 vscode）：
 *   EDITOR_SCHEME=goland pnpm dev   → 使用 GoLand 打开
 *   EDITOR_SCHEME=qoder pnpm dev    → 使用 Qoder 打开
 *   pnpm dev                        → 使用 VS Code 打开（默认）
 */
import type MarkdownIt from 'markdown-it'
import { resolve } from 'node:path'

// ─── 配置常量 ───────────────────────────────────────────────────────────────────
// GitHub 仓库地址（用于匹配 markdown 中的链接）
export const GITHUB_REPO = 'skyhe58/guide-go'
export const GITHUB_BASE_URL = `https://github.com/${GITHUB_REPO}/tree/main/`

// 编辑器协议配置
type EditorScheme = 'vscode' | 'qoder' | 'goland'

const EDITOR_SCHEME: EditorScheme =
  (process.env.EDITOR_SCHEME as EditorScheme) || 'vscode'

/**
 * 根据编辑器类型生成打开文件的 URL
 */
function buildEditorUrl(absolutePath: string): string {
  switch (EDITOR_SCHEME) {
    case 'goland':
      return `goland://open?file=${encodeURIComponent(absolutePath)}`
    case 'qoder':
      return `qoder://file${absolutePath}`
    case 'vscode':
    default:
      return `vscode://file${absolutePath}`
  }
}

// ─── markdown-it 插件：dev 模式下重写代码链接为编辑器协议 ───────────────────────────
export function codeLinksPlugin(md: MarkdownIt, projectRoot: string): void {
  const codeDir = resolve(projectRoot, 'code-examples')

  const defaultRender =
    md.renderer.rules.link_open ||
    ((tokens, idx, options, _env, self) => self.renderToken(tokens, idx, options))

  md.renderer.rules.link_open = (tokens, idx, options, env, self) => {
    const token = tokens[idx]
    const hrefIndex = token.attrIndex('href')

    if (hrefIndex >= 0) {
      const href = token.attrs![hrefIndex][1]
      const localPath = extractCodePath(href)

      if (localPath) {
        // 构建绝对路径并生成编辑器 URL
        const absolutePath = resolve(codeDir, localPath)
        token.attrs![hrefIndex][1] = buildEditorUrl(absolutePath)
      }
    }

    return defaultRender(tokens, idx, options, env, self)
  }
}

/**
 * 从 GitHub URL 中提取 code-examples/ 之后的路径
 * 支持格式：
 *   - https://github.com/skyhe58/guide-go/tree/main/code-examples/...
 */
function extractCodePath(href: string): string | null {
  if (!href.includes('github.com')) return null

  const marker = 'code-examples/'
  const markerIdx = href.indexOf(marker)
  if (markerIdx < 0) return null

  const path = href.slice(markerIdx + marker.length)
  if (!path) return null

  // 安全检查：防止目录遍历
  if (path.includes('..') || path.startsWith('/')) return null

  return path
}

/**
 * VitePress 代码链接环境动态切换插件
 *
 * - pnpm build: 保持 GitHub URL 原样输出
 * - pnpm dev:   重写为本地路径 + Vite 中间件提供文件服务
 */
import type MarkdownIt from 'markdown-it'
import { existsSync, readdirSync, statSync, readFileSync } from 'node:fs'
import { resolve, join, normalize, relative } from 'node:path'
import type { Plugin } from 'vite'

// ─── 配置常量 ───────────────────────────────────────────────────────────────────
// GitHub 仓库地址（用于匹配 markdown 中的链接）
export const GITHUB_REPO = 'skyhe58/guide-go'
export const GITHUB_BASE_URL = `https://github.com/${GITHUB_REPO}/tree/main/`
// 本地代码目录前缀
export const LOCAL_CODE_PREFIX = '/code-examples/'

// ─── markdown-it 插件：dev 模式下重写代码链接 ──────────────────────────────────────
export function codeLinksPlugin(md: MarkdownIt): void {
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
        // 重写为本地路径
        token.attrs![hrefIndex][1] = LOCAL_CODE_PREFIX + localPath
        // 新标签页打开，方便文档和代码对照
        token.attrSet('target', '_blank')
        token.attrSet('rel', 'noopener noreferrer')
      }
    }

    return defaultRender(tokens, idx, options, env, self)
  }
}

/**
 * 从 GitHub URL 中提取 code-examples/ 之后的路径
 * 支持多种格式：
 *   - https://github.com/skyhe58/guide-go/tree/main/code-examples/...
 *   - https://github.com/your-repo/code-examples/...
 *   - https://github.com/ (忽略，无有效路径)
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

// ─── Vite 插件：dev 模式下提供 code-examples 文件服务 ────────────────────────────
export function serveCodeExamples(projectRoot: string): Plugin {
  const codeDir = resolve(projectRoot, 'code-examples')

  return {
    name: 'serve-code-examples',
    configureServer(server) {
      server.middlewares.use((req, res, next) => {
        const url = req.url || ''
        if (!url.startsWith(LOCAL_CODE_PREFIX)) return next()

        // 提取请求路径
        const relativePath = decodeURIComponent(url.slice(LOCAL_CODE_PREFIX.length))

        // 安全检查
        if (relativePath.includes('..')) {
          res.statusCode = 403
          res.end('Forbidden')
          return
        }

        const fullPath = normalize(join(codeDir, relativePath))

        // 确保不超出 code-examples 目录
        if (!fullPath.startsWith(codeDir)) {
          res.statusCode = 403
          res.end('Forbidden')
          return
        }

        if (!existsSync(fullPath)) {
          res.statusCode = 404
          res.end(`Not found: ${relativePath}`)
          return
        }

        const stat = statSync(fullPath)

        if (stat.isDirectory()) {
          // 目录：返回文件列表 HTML
          const files = readdirSync(fullPath)
          const listItems = files
            .map((f) => {
              const fStat = statSync(join(fullPath, f))
              const suffix = fStat.isDirectory() ? '/' : ''
              return `<li><a href="${LOCAL_CODE_PREFIX}${relativePath}${relativePath.endsWith('/') ? '' : '/'}${f}${suffix}">${f}${suffix}</a></li>`
            })
            .join('\n')

          res.setHeader('Content-Type', 'text/html; charset=utf-8')
          res.end(`<!DOCTYPE html>
<html><head><meta charset="utf-8"><title>${relativePath}</title>
<style>body{font-family:monospace;padding:2rem;background:#1a1a1a;color:#e0e0e0}
a{color:#58a6ff;text-decoration:none}a:hover{text-decoration:underline}
li{margin:0.3rem 0}</style></head>
<body><h2>📁 code-examples/${relativePath}</h2><ul>${listItems}</ul></body></html>`)
        } else {
          // 文件：直接返回源码内容
          const content = readFileSync(fullPath, 'utf-8')
          const ext = fullPath.split('.').pop() || 'txt'
          const mimeMap: Record<string, string> = {
            go: 'text/x-go',
            mod: 'text/plain',
            sum: 'text/plain',
            ts: 'text/typescript',
            js: 'text/javascript',
            json: 'application/json',
            yaml: 'text/yaml',
            yml: 'text/yaml',
            toml: 'text/plain',
            md: 'text/markdown',
            txt: 'text/plain',
            sh: 'text/x-shellscript',
            dockerfile: 'text/plain',
          }
          const mime = mimeMap[ext.toLowerCase()] || 'text/plain'
          res.setHeader('Content-Type', `${mime}; charset=utf-8`)
          res.end(content)
        }
      })
    },
  }
}

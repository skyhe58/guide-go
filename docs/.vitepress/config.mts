import { defineConfig } from 'vitepress'
import { withMermaid } from 'vitepress-plugin-mermaid'
import { sidebar } from './sidebar.mts'

export default withMermaid(
  defineConfig({
    title: 'Go 知识库',
    description: 'Go 从入门到精通 — 面向中文开发者的系统学习知识库',
    lang: 'zh-CN',

    // 启用暗色模式
    appearance: true,

    // 构建时检测死链（暂时忽略）
    ignoreDeadLinks: true,

    themeConfig: {
      // 顶部导航
      nav: [
        { text: '首页', link: '/' },
        { text: '学习路径', link: '/learning-paths/beginner' },
        { text: '面试', link: '/interview/knowledge-map' },
        {
          text: '知识模块',
          items: [
            { text: '语言核心', link: '/1-go-core/1.1-go-basics/' },
            { text: 'Web 与数据', link: '/2-web-data/2.1-web-framework/' },
            { text: '微服务与云原生', link: '/3-microservice/3.1-microservice/' },
            { text: '分布式与架构', link: '/4-distributed/4.1-distributed/' },
            { text: '运维与部署', link: '/5-devops/5.1-cicd/' },
          ],
        },
        { text: 'GoBlog 实战', link: '/6-fullstack-project/' },
      ],

      // 侧边栏
      sidebar,

      // 本地搜索（支持中文）
      search: {
        provider: 'local',
        options: {
          translations: {
            button: {
              buttonText: '搜索文档',
              buttonAriaLabel: '搜索文档',
            },
            modal: {
              noResultsText: '无法找到相关结果',
              resetButtonTitle: '清除查询条件',
              footer: {
                selectText: '选择',
                navigateText: '切换',
                closeText: '关闭',
              },
            },
          },
        },
      },

      // 社交链接
      socialLinks: [
        { icon: 'github', link: 'https://github.com/your-username/guide-go' },
      ],

      // 页脚
      footer: {
        message: '基于 MIT 许可发布',
        copyright: 'Copyright © 2025 Go 知识库',
      },

      // 文档页脚导航
      docFooter: {
        prev: '上一页',
        next: '下一页',
      },

      // 大纲标题
      outline: {
        label: '页面导航',
        level: [2, 3],
      },

      // 最后更新时间
      lastUpdated: {
        text: '最后更新于',
      },

      // 返回顶部
      returnToTopLabel: '回到顶部',

      // 侧边栏菜单标签（移动端）
      sidebarMenuLabel: '菜单',

      // 暗色模式切换标签
      darkModeSwitchLabel: '主题',
    },

    // Markdown 配置
    markdown: {
      lineNumbers: true,
    },

    // 最后更新时间
    lastUpdated: true,
  })
)

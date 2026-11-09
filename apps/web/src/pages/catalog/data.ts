/**
 * Static catalog content for the 专家 / 技能 page.
 *
 * This page is deliberately render-only: the entries below are the intended
 * product showcase (what a Dongfeng tenant gets out of the box), not records
 * read from the backend. Nothing here is fetched, installed or executed — it
 * exists so people can see the intended lineup before the real catalog is
 * wired up. When the backend grows a real catalog endpoint, this file is the
 * single place the page's content comes from, so swapping in a request means
 * replacing `experts` / `skills` below and leaving the components untouched.
 */

export interface CatalogExpert {
  id: string
  /** Display name. The trailing qualifier (运营 / 用户) is part of the name on purpose. */
  name: string
  /** Short role summary, one line. */
  summary: string
  /** Longer description shown under the summary. */
  description?: string
  /** Companion/affiliation label rendered next to the avatar. */
  owner: string
  /** "MM-DD HH:mm 创建" line, kept as literal text — these are showcase entries. */
  createdAt: string
  /** Model badge, e.g. "openai/gpt-5" or "DeepSeek-V4". */
  model: string
  /** Accent used for the generated avatar tile; a token, never a raw hex. */
  accent: 'purple' | 'green' | 'blue' | 'orange' | 'red' | 'teal'
  /** Lucide icon name key resolved in the component. */
  icon: string
  /** Optional capability chips (e.g. 运维智能体). */
  tags?: string[]
  /**
   * Skill ids this expert carries by default. Adding an expert to a task means
   * inheriting THIS list — the product direction is "an agent picks one expert
   * to execute, and that expert decides which skills it brings". Ids resolve
   * against `catalogSkills`.
   */
  skills: string[]
  /** System prompt shown in the 应用选择 preview. */
  systemPrompt: string
}

export interface CatalogSkill {
  id: string
  name: string
  /** One-line description shown under the name, clamped to two lines. */
  description: string
  /** Version tag, e.g. "v1.0.0". */
  version: string
  /** Published date, literal showcase text. */
  publishedAt: string
  /** Grouping tab this skill belongs to. */
  group: 'builtin' | 'personal'
  /** Emoji-ish glyph shown in the tile. */
  icon: string
  accent: 'purple' | 'green' | 'blue' | 'orange' | 'red' | 'teal'
}

export const catalogExperts: CatalogExpert[] = [
  {
    id: 'zhiying-ops',
    name: '智算运营助理',
    summary: '智算资源与任务调度助手',
    owner: '重庆昇瑞',
    createdAt: '11-05 14:49 创建',
    model: 'openai/gpt-5',
    accent: 'purple',
    icon: 'Cpu',
    tags: ['运维智能体'],
    skills: ['create-job', 'xlsx', 'pdf'],
    systemPrompt: '你是东风集团智算运营助理，负责智算资源与任务调度。回答前先确认集群、队列和配额，再给出可执行的调度建议；涉及变更时给出回滚方案。',
  },
  {
    id: 'compliance-ops',
    name: '智驾合规运营小助理',
    summary: '智驾合规运营辅助，同时服务运营侧与用户侧',
    owner: '重庆昇瑞',
    createdAt: '10-21 17:08 创建',
    model: 'openai/gpt-5',
    accent: 'green',
    icon: 'ShieldCheck',
    skills: ['docx', 'doc-coauthoring', 'humanizer', 'pdf'],
    systemPrompt: '你是智驾合规运营小助理，同时服务运营侧与用户侧。涉及合规结论时必须引用具体条款来源，材料输出保持公文口吻。',
  },
  {
    id: 'scenario-ops',
    name: '500+场景运营经理',
    summary: '场景策划与智能体运营，服务东风集团各大区用户',
    owner: '重庆昇瑞',
    createdAt: '10-22 14:56 创建',
    model: 'openai/gpt-5',
    accent: 'orange',
    icon: 'Target',
    tags: ['场景策划'],
    skills: ['fde-elicitation', 'asp-ppt', 'pptx'],
    systemPrompt: '你是东风集团场景运营经理，负责 500+ 场景的策划与智能体运营。先把业务目标拆成可验证的场景假设，再产出方案与汇报材料。',
  },
  {
    id: 'qt-platform-ops',
    name: '擎天AI中台运营专家',
    summary: '擎天 AI 中台运营专家',
    owner: '重庆昇瑞',
    createdAt: '10-16 14:32 创建',
    model: 'openai/gpt-5',
    accent: 'teal',
    icon: 'Bot',
    skills: ['qingtian-creator', 'skill-creator', 'playwright-cli'],
    systemPrompt: '你是擎天 AI 中台运营专家，熟悉中台的智能体、技能与工作空间模型。回答优先复用中台既有能力，必要时给出新建智能体的配置草稿。',
  },
]

export const catalogSkills: CatalogSkill[] = [
  {
    id: 'fde-elicitation',
    name: 'fde-elicitation',
    description: '结构化（渐进式）需求撮合：把业务方「想要什么」拆成可验证的假设，先对齐再交付，避免做完才发现方向错了。',
    version: 'v1.0.0',
    publishedAt: '2026-10-31 13:09:21',
    group: 'builtin',
    icon: '🧩',
    accent: 'purple',
  },
  {
    id: 'qingtian-creator',
    name: 'qingtian-creator',
    description: '在擎天 AI 中台「提及擎天」时，创建新智能体并生成系统提示词、默认配置和可用工具清单。',
    version: 'v1.0.0',
    publishedAt: '2026-07-17 17:12:53',
    group: 'builtin',
    icon: '⚙️',
    accent: 'teal',
  },
  {
    id: 'playwright-cli',
    name: 'playwright-cli',
    description: 'Playwright 浏览器自动化：网页自动化、表单填写、截图、数据抓取，需要交互动效与端到端验证时使用。',
    version: 'v1.0.0',
    publishedAt: '2026-05-17 11:02:48',
    group: 'builtin',
    icon: '🎭',
    accent: 'green',
  },
  {
    id: 'pdf',
    name: 'pdf',
    description: 'PDF 全能处理：读取、提取文本与表格、合并拆分、旋转、加密解密、填表。把 PDF 当普通文档编辑。',
    version: 'v1.0.0',
    publishedAt: '2026-05-17 10:08:41',
    group: 'builtin',
    icon: '📄',
    accent: 'red',
  },
  {
    id: 'humanizer',
    name: 'humanizer',
    description: '去 AI 味：把生硬的模型腔改写成自然、有节奏的中文表达，去掉「首先/其次/综上所述」这类套话。',
    version: 'v1.0.0',
    publishedAt: '2026-06-17 11:02:12',
    group: 'builtin',
    icon: '✍️',
    accent: 'orange',
  },
  {
    id: 'docx',
    name: 'docx',
    description: 'Word 文档处理：创作、读取、编辑 .docx，保留样式、批注和修订，支持结构化长文排版。',
    version: 'v1.0.0',
    publishedAt: '2026-06-17 10:12:11',
    group: 'builtin',
    icon: '📝',
    accent: 'blue',
  },
  {
    id: 'doc-coauthoring',
    name: 'doc-coauthoring',
    description: '文档协作：把零散想法整理成可评审的结构化文档，生成大纲、拆章节、按读者调整详略。',
    version: 'v1.0.0',
    publishedAt: '2026-05-17 10:11:46',
    group: 'builtin',
    icon: '📚',
    accent: 'green',
  },
  {
    id: 'create-job',
    name: 'create-job',
    description: '创建定时任务：用 CRON 表达式定义触发时间，配置执行动作与失败重试，让重复工作自动跑起来。',
    version: 'v1.0.0',
    publishedAt: '2026-06-17 11:02:12',
    group: 'builtin',
    icon: '⏰',
    accent: 'purple',
  },
  {
    id: 'asp-ppt',
    name: 'asp-ppt',
    description: 'HTML 转 PPT：把网页、Markdown 或结构化内容渲染成审美在线的演示文稿，支持主题与母版。',
    version: 'v1.0.0',
    publishedAt: '2026-06-17 11:02:12',
    group: 'builtin',
    icon: '📊',
    accent: 'red',
  },
  {
    id: 'pptx',
    name: 'pptx',
    description: 'PowerPoint 处理：创建、编辑 .pptx，支持占位符、演讲者备注、图表、图片和版式复用。',
    version: 'v1.0.0',
    publishedAt: '2026-06-17 10:19:09',
    group: 'builtin',
    icon: '📈',
    accent: 'orange',
  },
  {
    id: 'skill-creator',
    name: 'skill-creator',
    description: '创建新技能：写 SKILL.md、定义触发条件与工具，把一次性的做法沉淀成团队可复用的能力。',
    version: 'v1.0.0',
    publishedAt: '2026-06-17 10:09:41',
    group: 'builtin',
    icon: '🛠️',
    accent: 'teal',
  },
  {
    id: 'xlsx',
    name: 'xlsx',
    description: 'Excel 处理：读取写入单元格、公式、透视表、图表，支持多工作表与 CSV 互转。',
    version: 'v1.0.0',
    publishedAt: '2026-06-17 10:10:20',
    group: 'builtin',
    icon: '📉',
    accent: 'blue',
  },
]

import type { ChatExample } from './types'

/**
 * Built-in usage examples. This is the single place example content lives;
 * UI reads it through useChatExamplesQuery, which is also where a backend
 * source would plug in (with this list as the offline fallback).
 *
 * The content is written for an internal enterprise deployment (东风集团):
 * routine office work, reporting, data and operational analysis. It must not
 * point at public developer services or consumer use cases — the platform is
 * self-hosted on an intranet, where those either do not resolve or are not
 * appropriate to suggest.
 *
 * Every example is offered on the welcome strip (three drawn at random per
 * new chat) and in the gallery. Only capability-free examples may be offered
 * during onboarding, because a freshly created assistant has nothing
 * configured yet.
 */
export const BUILTIN_CHAT_EXAMPLES: readonly ChatExample[] = [
  {
    id: 'meeting-minutes',
    category: 'team',
    icon: 'notebook-pen',
    priority: 99,
    requires: [],
    surfaces: ['welcome', 'gallery'],
    title: { en: 'Turn notes into minutes', zh: '把会议记录整理成纪要', ja: '会議メモを議事録に' },
    prompt: {
      en: 'Turn the meeting notes below into minutes: decisions made, questions left open, and next actions with an owner and a due date for each. Flag anything that was discussed but never settled. Here are the notes:\n\n',
      zh: '把下面的会议记录整理成会议纪要：已经确定的事项、还悬着的问题、下一步动作（含责任人和时间节点）。讨论过但没有结论的，单独标出来提醒我。原始记录如下：\n\n',
      ja: '以下の会議メモを議事録にまとめてください。決定事項、未解決の論点、次のアクション（担当者と期限つき）を整理し、議論されたが結論が出ていないものは別途指摘してください。元のメモはこちらです：\n\n',
    },
  },
  {
    id: 'weekly-report',
    category: 'writing',
    icon: 'pen-line',
    priority: 98,
    requires: [],
    surfaces: ['welcome', 'gallery'],
    title: { en: 'Draft my weekly report', zh: '写这周的周报', ja: '今週の週報を書く' },
    prompt: {
      en: 'Write this week\'s report for me. Ask what I finished, what is still in progress and what is blocked, then turn it into the usual structure: this week\'s results, open items with reasons, next week\'s plan. Keep it factual — no padding.\n\n',
      zh: '帮我写这周的周报。先问清楚我完成了什么、哪些还在推进、哪些被卡住了，然后整理成固定结构：本周完成情况、未完成事项及原因、下周计划。实事求是，不要空话。\n\n',
      ja: '今週の週報を書いてください。まず完了したこと、進行中のこと、止まっていることを質問し、いつもの構成（今週の実績、未完了の項目と理由、来週の計画）にまとめてください。事実ベースで、飾りの文章は不要です。\n\n',
    },
  },
  {
    id: 'dashboard-analysis',
    category: 'research',
    icon: 'chart-pie',
    priority: 97,
    requires: ['workspace'],
    surfaces: ['welcome', 'gallery'],
    title: { en: 'Explain a data swing', zh: '分析一期数据的异动', ja: 'データの変動を分析' },
    prompt: {
      en: 'I exported a period of dashboard data into the workspace. Read it, compare each metric with the previous period, and list the ones that moved more than 10% — with the units that changed most and the likely cause for each. Put the numbers in a table I can paste into my report.\n\n',
      zh: '我把一批看板数据导到工作区了。读一下，把每个指标和上期对比，列出变动超过 10% 的项，并指出是哪些单位引起的、可能的原因是什么。数字整理成表格，我要直接贴进汇报材料：\n\n',
      ja: 'ダッシュボードのデータをワークスペースに書き出しました。読み込んで、各指標を前期と比較し、10% 以上変動した項目を挙げてください。どの単位の変化が大きいか、考えられる原因も添えてください。数値は報告資料に貼れる表にまとめてください：\n\n',
    },
  },
  {
    id: 'report-material',
    category: 'writing',
    icon: 'newspaper',
    priority: 96,
    requires: [],
    surfaces: ['welcome', 'gallery'],
    title: { en: 'Draft a briefing paper', zh: '起草一份汇报材料', ja: '報告資料の下書き' },
    prompt: {
      en: 'Help me draft a briefing paper for leadership. Ask me the topic, the audience, the length limit and what decision I want out of it first. Then write it in plain official style: background, current state with numbers, problems, proposed next steps. Avoid slogans and empty modifiers.\n\n',
      zh: '帮我起草一份向上汇报的材料。先问清楚主题、汇报对象、篇幅要求和希望领导拍板的事项，然后按公文口径写：背景、现状（带数据）、存在问题、下一步建议。不要口号，不要空泛的形容词。\n\n',
      ja: '上司への報告資料の下書きを手伝ってください。まずテーマ、報告先、分量、判断してほしい事項を質問し、その後は公文書の書き方で、背景・現状（数値つき）・課題・今後の提案の順に書いてください。スローガンや空虚な修飾語は不要です。\n\n',
    },
  },
  {
    id: 'policy-qa',
    category: 'research',
    icon: 'book-open',
    priority: 95,
    requires: [],
    surfaces: ['welcome', 'gallery'],
    title: { en: 'Answer from our rules', zh: '按制度文件回答问题', ja: '規程に基づいて回答' },
    prompt: {
      en: 'Answer strictly from the internal policy documents I give you — no outside knowledge, no guessing. For every conclusion, quote the clause and its number. If the documents do not cover it, say so and tell me which department to ask. My question:\n\n',
      zh: '只根据我提供的内部制度文件回答，不要用外部知识，也不要猜。每个结论都要引用条款号和原文。文件里没写到的，直接说没写，并告诉我该问哪个部门。我的问题：\n\n',
      ja: '私が渡す社内規程の文書だけに基づいて回答してください。外部知識や推測は使わないでください。結論ごとに条項番号と原文を引用します。文書に書かれていない場合は、その旨と、どの部署に確認すべきかを教えてください。質問はこちらです：\n\n',
    },
  },
  {
    id: 'excel-cleanup',
    category: 'dev',
    icon: 'square-terminal',
    priority: 94,
    requires: ['workspace'],
    surfaces: ['welcome', 'gallery'],
    title: { en: 'Clean up a spreadsheet', zh: '清理一张台账表', ja: '台帳のデータを整える' },
    prompt: {
      en: 'There is an Excel ledger in the workspace. Normalize it: unify the date format, trim whitespace, fix the inconsistent unit names, and find duplicate rows. Save the cleaned file next to the original and give me a short list of what you changed and how many rows were affected.\n\n',
      zh: '工作区里有一张台账表。帮我做数据清理：统一日期格式、去掉多余空格、把写法不一致的单位名称归并、找出重复行。清理后的文件另存一份，然后告诉我改了哪些地方、各影响多少行：\n\n',
      ja: 'ワークスペースに台帳の Excel があります。データを整えてください。日付形式の統一、余分な空白の削除、表記ゆれのある部署名の名寄せ、重複行の検出を行います。整えたファイルは元と別名で保存し、どこをどう直したか、影響行数とともに教えてください：\n\n',
    },
  },
  {
    id: 'week-schedule',
    category: 'writing',
    icon: 'calendar-days',
    priority: 93,
    requires: [],
    surfaces: ['welcome', 'gallery'],
    title: { en: 'Sort out next week', zh: '排一下下周的安排', ja: '来週の予定を組み立てる' },
    prompt: {
      en: 'Help me arrange next week. Ask about my fixed meetings, the deadlines I cannot move, and the one or two things that actually must get done. Then give me a day-by-day plan with realistic blocks and mark where I have no buffer left.\n\n',
      zh: '帮我排下周的安排。先问我有哪些固定会议、哪些截止时间不能动、哪一两件事这周必须出结果，然后给一份按天排的计划，时间块要现实一点，并标出哪里已经没有缓冲了。\n\n',
      ja: '来週の予定を組み立ててください。固定の会議、動かせない締め切り、今週必ず結果を出す一〜二件を質問し、そのうえで日ごとの計画を提案してください。時間枠は現実的にし、余裕がなくなっている箇所を指摘してください。\n\n',
    },
  },
  {
    id: 'mail-digest',
    category: 'automation',
    icon: 'mail',
    priority: 90,
    requires: ['email', 'schedule'],
    surfaces: ['welcome', 'gallery'],
    title: { en: 'Morning inbox digest', zh: '早间邮件摘要', ja: '朝のメールダイジェスト' },
    prompt: {
      en: 'Every workday at 8, go through the mail that arrived overnight, group it by urgency (needs my reply today / this week / informational), and send me one short digest. Do not send anything on my behalf.\n\n',
      zh: '每个工作日早上 8 点，把夜里收到的邮件过一遍，按紧急程度分组（今天必须回 / 本周回 / 仅知会），发我一份简短摘要。不要替我给任何人发信。\n\n',
      ja: '平日の毎朝 8 時に、夜間に届いたメールを確認し、緊急度で分類して（今日返信が必要／今週中／共有のみ）短いダイジェストを送ってください。代理で送信はしないでください。\n\n',
    },
  },
  {
    id: 'weekly-reminder',
    category: 'team',
    icon: 'users',
    priority: 89,
    requires: ['schedule', 'channel'],
    surfaces: ['welcome', 'gallery'],
    title: { en: 'Chase the weekly reports', zh: '催收周报', ja: '週報の提出督促' },
    prompt: {
      en: 'Every Friday at 4 pm, remind the team group chat to submit their weekly reports. At 7 pm, send me the list of who has not — as a table with name and how many weeks in a row they are late.\n\n',
      zh: '每周五下午 4 点在团队群里提醒大家交周报，晚上 7 点把还没交的人整理成名单发我，标上连续迟交了几周。\n\n',
      ja: '毎週金曜 16 時にチームのグループチャットで週報提出のリマインドをし、19 時に未提出者の一覧を、連続遅延週数つきで送ってください。\n\n',
    },
  },
  {
    id: 'industry-brief',
    category: 'research',
    icon: 'search',
    priority: 88,
    requires: ['web_search'],
    surfaces: ['welcome', 'gallery'],
    title: { en: 'Industry policy brief', zh: '行业政策简报', ja: '業界政策のブリーフ' },
    prompt: {
      en: 'Put together a one-page brief on recent policy moves that affect our business. Only use sources you can cite with a title, issuing body and date. Separate confirmed facts from my own judgement, and finish with three things we should prepare for.\n\n',
      zh: '整理一份一页纸的行业政策简报，聚焦最近和我们业务相关的政策动向。只用能标出来源、发文单位和日期的材料。事实和我的判断要分开写，最后给出三条我们该提前准备的事：\n\n',
      ja: 'わが社の事業に関係する直近の政策動向について、1 ページのブリーフを作ってください。出典・発行機関・日付を示せる資料だけを使ってください。事実と私の判断は分けて書き、最後に準備すべき三点を挙げてください：\n\n',
    },
  },
  {
    id: 'requirement-doc',
    category: 'writing',
    icon: 'file-text',
    priority: 87,
    requires: [],
    surfaces: ['welcome', 'gallery'],
    title: { en: 'Shape a rough request', zh: '把口头需求写成文档', ja: '口頭の要望を文書化' },
    prompt: {
      en: 'Turn the rough request below into a proper requirement note: who needs it and why, what happens today, what should happen instead, the acceptance criteria, and the open questions we must settle before development. Point out anything ambiguous instead of filling it in yourself.\n\n',
      zh: '把下面这段口头需求写成正规的需求说明：谁提的、要解决什么问题、现在怎么做、希望改成什么样、验收标准是什么，以及开发前必须先定下来的问题。有含糊的地方直接指出来，不要自己替我补完：\n\n',
      ja: '以下の口頭の要望を要件定義メモにしてください。誰が何のために必要か、現状の運用、望ましい姿、受け入れ条件、開発前に決めるべき未確定事項を整理します。曖昧な点は勝手に補完せず、指摘してください：\n\n',
    },
  },
  {
    id: 'translate-doc',
    category: 'writing',
    icon: 'globe',
    priority: 86,
    requires: [],
    surfaces: ['welcome', 'gallery'],
    title: { en: 'Translate a document', zh: '翻译一份文档', ja: '文書を翻訳' },
    prompt: {
      en: 'Translate the document below. Keep the terminology consistent with how we already name things — do not invent new translations. Preserve headings, numbering and tables. Afterwards list the terms you were unsure about so I can confirm them.\n\n',
      zh: '翻译下面这份文档。术语沿用我们已有的叫法，不要自己新造译名。保留标题层级、编号和表格结构。翻译完把你不确定的地方列出来，我来确认：\n\n',
      ja: '以下の文書を翻訳してください。用語は既存の呼び方に合わせ、新しい訳語を作らないでください。見出しの階層、番号、表の構造は維持します。翻訳後に、判断に迷った用語を列挙して確認を求めてください：\n\n',
    },
  },
  {
    id: 'contract-review',
    category: 'research',
    icon: 'target',
    priority: 60,
    requires: ['workspace'],
    surfaces: ['welcome', 'gallery'],
    title: { en: 'Pull the key terms', zh: '提取合同关键条款', ja: '契約の主要条項を抽出' },
    prompt: {
      en: 'Read the contract in the workspace and pull out what matters: scope, term and renewal, payment terms, liability, termination and any clause that is unusually one-sided. Quote the original text for each point and flag what needs legal review. This is a first pass, not a legal opinion.\n\n',
      zh: '读一下工作区里的这份合同，把关键条款提取出来：范围、期限与续签、付款、违约责任、解除条件，以及明显单方面不利的条款。每一条都附原文，并标出哪些必须走法务。这只是初筛，不代替法务意见：\n\n',
      ja: 'ワークスペース内の契約書を読み、重要な条項を抜き出してください。対象範囲、期間と更新、支払条件、責任、解除条件、および明らかに一方に不利な条項です。各項目に原文を添え、法務確認が必要な箇所を示してください。これは一次確認であり、法務意見ではありません：\n\n',
    },
  },
  {
    id: 'knowledge-sort',
    category: 'automation',
    icon: 'book-open',
    priority: 55,
    requires: ['workspace'],
    surfaces: ['gallery'],
    title: { en: 'Sort a document pile', zh: '梳理一批文档', ja: '文書群を整理' },
    prompt: {
      en: 'The workspace has a pile of files with unclear names. Group them by topic, propose a folder structure and a naming rule, then rename and move them. Give me the mapping table first and wait for my confirmation before touching anything.\n\n',
      zh: '工作区里有一堆文件名很乱的文件。先按主题归类，提出一套目录结构和命名规则，然后把它们改名归位。动手之前先把对照表给我，我确认了再改：\n\n',
      ja: 'ワークスペースに名前が整理されていないファイルが多数あります。テーマ別に分類し、フォルダ構成と命名ルールを提案したうえで、改名と移動を行ってください。作業前にまず対応表を出し、私の確認を待ってください：\n\n',
    },
  },
  {
    id: 'polish-writing',
    category: 'writing',
    icon: 'pen-line',
    priority: 30,
    requires: [],
    surfaces: ['welcome', 'gallery', 'onboarding'],
    title: { en: 'Polish my writing', zh: '润色一段文字', ja: '文章を整える' },
    prompt: {
      en: 'Make this text clearer and more concise while keeping my tone and every fact intact. Afterwards explain the main changes:\n\n',
      zh: '帮我把下面这段文字改得更清楚、更简洁，保留我的语气，事实一处都不要动。最后说明主要改了哪里：\n\n',
      ja: '次の文章を、私の口調と事実関係を保ったまま、より分かりやすく簡潔にしてください。最後に主な変更点を説明してください：\n\n',
    },
  },
  {
    id: 'plan-my-week',
    category: 'writing',
    icon: 'calendar-days',
    priority: 25,
    requires: [],
    surfaces: ['welcome', 'gallery', 'onboarding'],
    title: { en: 'Plan my week', zh: '安排这一周', ja: '今週の計画' },
    prompt: {
      en: 'Help me plan this week. Ask me about my deadlines and fixed commitments first, then propose a realistic day-by-day schedule.\n\n',
      zh: '帮我安排这一周。先问我有哪些截止时间和固定安排，再给出一份切实可行的每日计划。\n\n',
      ja: '今週の計画を手伝ってください。まず締め切りや固定の予定を質問し、そのうえで無理のない日ごとのスケジュールを提案してください。\n\n',
    },
  },
  {
    id: 'explain-concept',
    category: 'research',
    icon: 'lightbulb',
    priority: 20,
    requires: [],
    surfaces: ['welcome', 'gallery', 'onboarding'],
    title: { en: 'Explain a concept', zh: '讲清楚一个概念', ja: '概念をわかりやすく' },
    prompt: {
      en: 'Explain a concept to me step by step, starting from what I already know. Ask me which concept and how familiar I am with it first.\n\n',
      zh: '一步步给我讲清楚一个概念，从我已经知道的部分讲起。先问我想了解哪个概念、我对它有多熟悉。\n\n',
      ja: 'ある概念を、私がすでに知っていることから順を追って説明してください。まず、どの概念か、どのくらい知っているかを質問してください。\n\n',
    },
  },
  {
    id: 'daily-review',
    category: 'memory',
    icon: 'trending-up',
    priority: 40,
    requires: [],
    surfaces: ['gallery'],
    title: { en: 'End-of-day wrap-up', zh: '下班前过一遍今天', ja: '退勤前の振り返り' },
    prompt: {
      en: 'Before I leave, run through today with me: what got done, what slipped and why, and what tomorrow\'s first task should be. Keep it short and put the unfinished items where I will see them tomorrow.\n\n',
      zh: '下班前陪我过一遍今天：做完了什么、什么没做成、原因是什么，明天早上第一件事该干什么。简短点说，没做完的放到明天能看到的地方。\n\n',
      ja: '退勤前に今日を一緒に振り返ってください。できたこと、できなかったこととその理由、明日の最初の作業。短くまとめ、終わらなかった項目は明日目につく場所に置いてください。\n\n',
    },
  },
]

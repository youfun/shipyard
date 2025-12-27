export const dict = {
  // Common
  common: {
    save: "保存",
    cancel: "取消",
    edit: "编辑",
    delete: "删除",
    add: "添加",
    loading: "加载中...",
    yes: "是",
    error: "错误",
    close: "关闭",
    actions: "操作",
    confirm_delete: "您确定要删除此项目吗？",
  },

  // Navigation
  nav: {
    admin_title: "管理后台",
    system_title: "管理系统",
    dashboard: "仪表板",
    applications: "应用管理",
    ssh_management: "SSH管理",
    settings: "设置",
    change_password: "修改密码",
    logout: "退出登录",
    version: "版本",
  },

  // Login Page
  login: {
    title: "管理员登录",
    username: "用户名",
    password: "密码",
    username_placeholder: "请输入用户名",
    password_placeholder: "请输入密码",
    login_button: "登录",
    logging_in: "登录中...",
    empty_fields_error: "用户名和密码不能为空",
    login_failed: "登录失败，请重试",
  },

  // Dashboard Page
  dashboard: {
    cards: {
      applications: "应用",
      deployed_applications: "已部署的应用",
      hosts: "主机",
      connected_hosts: "已连接的主机",
      deployments: "部署",
      total_deployments: "总部署次数",
    },
    recent_deployments: {
      title: "最近部署（20个）",
      no_deployments: "暂无部署记录",
      host: "主机",
      time: "部署时间",
      app_name: "应用名称",
      version: "版本号",
      git_commit: "Git提交",
      status: "状态",
      port: "端口",
    },
  },

  // Language
  language: {
    chinese: "中文",
    english: "English",
  },

  // SSH Management
  ssh: {
    title: "SSH管理",
    description: "管理远程SSH主机连接",
    add_host: "添加主机",
    edit_host: "编辑主机",
    name: "主机名称",
    address: "地址",
    user: "用户名",
    port: "端口",
    password: "密码",
    private_key: "私钥",
    no_hosts: "暂无SSH主机",
    confirm_delete: "确认删除",
    delete_warning: "确定要删除SSH主机 \"{name}\" 吗？此操作不可恢复。",
    name_placeholder: "请输入主机名称",
    address_placeholder: "IP地址或主机名",
    user_placeholder: "SSH用户名",
    password_placeholder: "SSH密码（可选）",
    private_key_placeholder: "SSH私钥内容（可选）",
  },

  // Change Password Page
  change_password: {
    title: "修改密码",
    current_password: "当前密码",
    new_password: "新密码",
    confirm_password: "确认新密码",
    current_password_placeholder: "请输入当前密码",
    new_password_placeholder: "请输入新密码",
    confirm_password_placeholder: "请再次输入新密码",
    change_button: "修改密码",
    changing: "修改中...",
    success_message: "密码修改成功",
    error_empty_fields: "当前密码和新密码不能为空",
    error_password_length: "新密码长度至少6位",
    error_password_mismatch: "新密码确认不匹配",
    error_change_failed: "修改密码失败，请重试",
  },

  // CLI Device Authorization
  cli_device_auth: {
    title: "设备授权确认",
    subtitle: "您正在授权一台新设备登录",
    confirm_info: "请确认以下信息是否与您发起登录的设备一致：",
    device_system: "设备系统",
    device_name: "设备名称",
    ip_address: "IP 地址",
    request_time: "请求时间",
    authorize_app: "授权应用",
    app_name: "OrbitCtl CLI Tool",
    confirm_button: "✅ 是我本人，确认登录",
    deny_button: "❌ 不是我，拒绝",
    authorizing: "授权中...",
    success_title: "授权成功！",
    success_message: "设备已成功授权，您现在可以关闭此页面。",
    auto_close_message: "页面将在 3 秒后自动关闭...",
    cancel_close: "取消并关闭",
  },

  // Setup Page
  setup: {
    title: "初始化设置",
    subtitle: "首次使用需要设置管理员账号",
    admin_username: "管理员用户名",
    username_placeholder: "请输入用户名",
    password: "密码",
    password_placeholder: "请输入密码",
    confirm_password: "确认密码",
    confirm_password_placeholder: "请再次输入密码",
    setup_button: "完成设置",
    setting_up: "设置中...",
    error_empty_fields: "用户名和密码不能为空",
    error_password_mismatch: "密码确认不匹配",
    error_password_length: "密码长度至少6位",
    error_setup_failed: "设置失败",
    error_network: "网络错误，请重试",
  },

  // Setup Guard
  setup_guard: {
    checking_status: "正在检查系统状态...",
  },

  // Security Settings Page
  security_settings: {
    description: "管理您的密码和双因素身份验证",
  },

  // System Settings
  system_settings: {
    title: "系统设置",
    description: "管理系统范围的设置",
    domain: "系统域名",
    domain_placeholder: "输入系统的域名",
    error_empty_domain: "域名不能为空",
    error_update_failed: "更新系统设置失败",
    success_message: "系统设置已成功更新",
  },

  // Application List Page
  app_list: {
    title: "应用列表",
    empty_title: "暂无应用",
    empty_description: "使用 CLI 创建并部署您的第一个应用",
    table_name: "名称",
    table_last_deployment: "最后部署",
    table_linked_host: "关联主机",
    table_actions: "操作",
    action_details: "详情",
  },

  // Application Detail Page
  app_detail: {
    loading: "加载中...",
    error_not_found: "应用未找到",
    action_back: "返回应用列表",
    breadcrumb_home: "首页",
    breadcrumb_applications: "应用",

    // Tabs
    tab_overview: "概览",
    tab_deployments: "部署历史",
    tab_environment: "环境变量",
    tab_domains: "域名",
    tab_tokens: "令牌",
    tab_settings: "设置",

    // Overview Tab
    overview_title: "应用信息",
    overview_name: "名称",
    overview_status: "状态",
    overview_domain: "域名",
    overview_target_port: "目标端口",
    overview_created: "创建时间",
    overview_last_deployed: "最后部署",
    overview_none: "-",

    // Instances
    instances_title: "运行实例",
    instance_host: "主机",
    instance_status: "状态",
    instance_port: "端口",
    instance_actions: "操作",
    action_start: "启动",
    action_stop: "停止",
    action_restart: "重启",
    action_view_logs: "查看日志",
    logs_title: "实例日志",
    deployment_logs_title: "部署日志",
    logs_lines: "行",
    logs_refresh: "刷新",
    logs_copy: "复制",
    logs_copied: "日志已复制到剪贴板",
    logs_copy_failed: "复制日志失败",
    logs_empty: "暂无日志",
    logs_error: "获取日志失败",
    logs_start_realtime: "实时日志",
    logs_stop_realtime: "停止实时",
    logs_stream_error: "无法连接日志流",

    // Deployments Tab
    deployments_empty: "暂无部署记录",
    deployments_version: "版本",
    deployments_status: "状态",
    deployments_host: "主机",
    deployments_port: "端口",
    deployments_created: "创建时间",

    // Environment Tab
    environment_empty: "暂无环境变量配置",
    environment_key: "键",
    environment_value: "值",
    environment_encrypted: "已加密",
    environment_yes: "是",
    environment_no: "否",
    environment_encrypted_placeholder: "输入新值以更新",
    environment_encrypted_warning: "此值已加密。留空以保持现有值。",

    // Tokens Tab
    tokens_title: "API 令牌",
    tokens_description: "管理用于 CLI 访问和部署的 API 令牌",
    tokens_create: "创建令牌",
    tokens_empty: "暂无 API 令牌",
    tokens_empty_description: "创建令牌以启用 CLI 访问",
    tokens_table_name: "名称",
    tokens_table_created: "创建时间",
    tokens_table_expires: "过期时间",
    tokens_table_last_used: "最后使用",
    tokens_table_actions: "操作",
    tokens_never: "永不过期",
    tokens_expired: "已过期",
    tokens_never_used: "从未使用",

    tokens_modal_create_title: "创建新令牌",
    tokens_modal_success_title: "令牌创建成功",
    tokens_warning_copy: "请确保现在复制此令牌。您将无法再次看到它！",
    tokens_label_token: "令牌",
    tokens_action_copy: "复制",
    tokens_action_done: "完成",
    tokens_label_name: "令牌名称",
    tokens_placeholder_name: "例如：CI/CD 流水线",
    tokens_hint_name: "用于识别此令牌的描述性名称",
    tokens_label_expiration: "过期时间（可选）",
    tokens_option_never: "永不过期",
    tokens_option_7days: "7 天",
    tokens_option_30days: "30 天",
    tokens_option_90days: "90 天",
    tokens_option_180days: "180 天",
    tokens_option_1year: "1 年",
    tokens_action_cancel: "取消",
    tokens_action_create: "创建令牌",
    tokens_copied: "令牌已复制到剪贴板",
    tokens_copy_failed: "复制令牌失败",

    tokens_delete_title: "删除令牌",
    tokens_delete_message: "确定要删除此令牌吗？此操作无法撤销，使用此令牌的任何服务都将失去访问权限。",
    tokens_delete_cancel: "取消",
    tokens_delete_confirm: "删除",

    // Settings Tab
    settings_title: "应用设置",
    settings_description: "在此配置应用设置",
    settings_label_name: "应用名称",
    settings_label_description: "描述",
  },

  tokens_development_notice: "🚧 功能开发中 - Token 将用于自动化部署（类似 GitHub Actions）",
}

export type Dictionary = typeof dict

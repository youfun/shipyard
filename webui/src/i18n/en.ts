export const dict = {
  // Common
  common: {
    save: "Save",
    cancel: "Cancel",
    edit: "Edit",
    delete: "Delete",
    add: "Add",
    loading: "Loading...",
    yes: "Yes",
    error: "Error",
    close: "Close",
    actions: "Actions",
    confirm_delete: "Are you sure you want to delete this item?",
  },

  // Navigation
  nav: {
    admin_title: "Admin Panel",
    system_title: "Management System",
    dashboard: "Dashboard",
    applications: "Applications",
    ssh_management: "SSH Management",
    settings: "Settings",
    change_password: "Change Password",
    logout: "Logout",
    version: "Version",
  },

  // Login Page
  login: {
    title: "Admin Login",
    username: "Username",
    password: "Password",
    username_placeholder: "Enter username",
    password_placeholder: "Enter password",
    login_button: "Login",
    logging_in: "Logging in...",
    empty_fields_error: "Username and password cannot be empty",
    login_failed: "Login failed, please try again",
  },

  // Dashboard Page
  dashboard: {
    cards: {
      applications: "Applications",
      deployed_applications: "Deployed Applications",
      hosts: "Hosts",
      connected_hosts: "Connected Hosts",
      deployments: "Deployments",
      total_deployments: "Total Deployments",
    },
    recent_deployments: {
      title: "Recent Deployments (20)",
      no_deployments: "No deployment records",
      host: "Host",
      time: "Deployment Time",
      app_name: "Application Name",
      version: "Version",
      git_commit: "Git Commit",
      status: "Status",
      port: "Port",
    },
  },

  // Language
  language: {
    chinese: "中文",
    english: "English",
  },

  // SSH Management
  ssh: {
    title: "SSH Management",
    description: "Manage remote SSH host connections",
    add_host: "Add Host",
    edit_host: "Edit Host",
    name: "Host Name",
    address: "Address",
    user: "Username",
    port: "Port",
    password: "Password",
    private_key: "Private Key",
    no_hosts: "No SSH hosts",
    confirm_delete: "Confirm Delete",
    delete_warning: "Are you sure you want to delete SSH host \"{name}\"? This action cannot be undone.",
    name_placeholder: "Enter host name",
    address_placeholder: "IP address or hostname",
    user_placeholder: "SSH username",
    password_placeholder: "SSH password (optional)",
    private_key_placeholder: "SSH private key content (optional)",
  },

  // Change Password Page
  change_password: {
    title: "Change Password",
    current_password: "Current Password",
    new_password: "New Password",
    confirm_password: "Confirm New Password",
    current_password_placeholder: "Enter current password",
    new_password_placeholder: "Enter new password",
    confirm_password_placeholder: "Enter new password again",
    change_button: "Change Password",
    changing: "Changing...",
    success_message: "Password changed successfully",
    error_empty_fields: "Current password and new password cannot be empty",
    error_password_length: "New password must be at least 6 characters",
    error_password_mismatch: "New password confirmation does not match",
    error_change_failed: "Failed to change password, please try again",
  },

  // CLI Device Authorization
  cli_device_auth: {
    title: "Device Authorization Confirmation",
    subtitle: "You are authorizing a new device to log in",
    confirm_info: "Please confirm if the following information matches the device you used to initiate the login:",
    device_system: "Device System",
    device_name: "Device Name",
    ip_address: "IP Address",
    request_time: "Request Time",
    authorize_app: "Authorize Application",
    app_name: "OrbitCtl CLI Tool",
    confirm_button: "✅ Yes, it's me. Confirm login",
    deny_button: "❌ No, it's not me. Deny",
    authorizing: "Authorizing...",
    success_title: "Authorization Successful!",
    success_message: "The device has been successfully authorized. You can now close this page.",
    auto_close_message: "This page will automatically close in 3 seconds...",
    cancel_close: "Cancel and Close",
  },

  // Setup Page
  setup: {
    title: "Initial Setup",
    subtitle: "First-time setup requires administrator account configuration",
    admin_username: "Administrator Username",
    username_placeholder: "Enter username",
    password: "Password",
    password_placeholder: "Enter password",
    confirm_password: "Confirm Password",
    confirm_password_placeholder: "Enter password again",
    setup_button: "Complete Setup",
    setting_up: "Setting up...",
    error_empty_fields: "Username and password cannot be empty",
    error_password_mismatch: "Password confirmation does not match",
    error_password_length: "Password must be at least 6 characters",
    error_setup_failed: "Setup failed",
    error_network: "Network error, please try again",
  },

  // Setup Guard
  setup_guard: {
    checking_status: "Checking system status...",
  },

  // Security Settings Page
  security_settings: {
    description: "Manage your password and two-factor authentication",
  },

  // System Settings
  system_settings: {
    title: "System Settings",
    description: "Manage system-wide settings",
    domain: "System Domain",
    domain_placeholder: "Enter system domain",
    error_empty_domain: "Domain cannot be empty",
    error_update_failed: "Failed to update system settings",
    success_message: "System settings updated successfully",
  },

  // Application List Page
  app_list: {
    title: "Applications",
    empty_title: "No Applications",
    empty_description: "Use the CLI to create and deploy your first application",
    table_name: "Name",
    table_last_deployment: "Last Deployment",
    table_linked_host: "Linked Host",
    table_actions: "Actions",
    action_details: "Details",
  },

  // Application Detail Page
  app_detail: {
    loading: "Loading...",
    error_not_found: "Application not found",
    action_back: "Back to Applications",
    breadcrumb_home: "Home",
    breadcrumb_applications: "Applications",

    // Tabs
    tab_overview: "Overview",
    tab_deployments: "Deployment History",
    tab_environment: "Environment Variables",
    tab_domains: "Domains",
    tab_tokens: "Tokens",
    tab_settings: "Settings",

    // Overview Tab
    overview_title: "Application Information",
    overview_name: "Name",
    overview_status: "Status",
    overview_domain: "Domain",
    overview_target_port: "Target Port",
    overview_created: "Created",
    overview_last_deployed: "Last Deployed",
    overview_none: "-",

    // Instances
    instances_title: "Running Instances",
    instance_host: "Host",
    instance_status: "Status",
    instance_port: "Port",
    instance_actions: "Actions",
    action_start: "Start",
    action_stop: "Stop",
    action_restart: "Restart",
    action_view_logs: "View Logs",
    logs_title: "Instance Logs",
    deployment_logs_title: "Deployment Logs",
    logs_lines: "lines",
    logs_refresh: "Refresh",
    logs_copy: "Copy",
    logs_copied: "Logs copied to clipboard",
    logs_copy_failed: "Failed to copy logs",
    logs_empty: "No logs available",
    logs_error: "Failed to fetch logs",
    logs_start_realtime: "Real-time Logs",
    logs_stop_realtime: "Stop Real-time",
    logs_stream_error: "Unable to connect to log stream",

    // Deployments Tab
    deployments_empty: "No deployment records",
    deployments_version: "Version",
    deployments_status: "Status",
    deployments_host: "Host",
    deployments_port: "Port",
    deployments_created: "Created",

    // Environment Tab
    environment_empty: "No environment variables configured",
    environment_key: "Key",
    environment_value: "Value",
    environment_encrypted: "Encrypted",
    environment_yes: "Yes",
    environment_no: "No",
    environment_encrypted_placeholder: "Enter new value to update",
    environment_encrypted_warning: "This value is encrypted. Leave blank to keep existing value.",

    // Tokens Tab
    tokens_title: "API Tokens",
    tokens_description: "Manage API tokens for CLI access and deployments",
    tokens_create: "Create Token",
    tokens_empty: "No API tokens",
    tokens_empty_description: "Create a token to enable CLI access",
    tokens_table_name: "Name",
    tokens_table_created: "Created",
    tokens_table_expires: "Expires",
    tokens_table_last_used: "Last Used",
    tokens_table_actions: "Actions",
    tokens_never: "Never expires",
    tokens_expired: "Expired",
    tokens_never_used: "Never used",

    tokens_modal_create_title: "Create New Token",
    tokens_modal_success_title: "Token Created Successfully",
    tokens_warning_copy: "Make sure to copy this token now. You won't be able to see it again!",
    tokens_label_token: "Token",
    tokens_action_copy: "Copy",
    tokens_action_done: "Done",
    tokens_label_name: "Token Name",
    tokens_placeholder_name: "e.g., CI/CD Pipeline",
    tokens_hint_name: "A descriptive name to identify this token",
    tokens_label_expiration: "Expiration (optional)",
    tokens_option_never: "Never expires",
    tokens_option_7days: "7 days",
    tokens_option_30days: "30 days",
    tokens_option_90days: "90 days",
    tokens_option_180days: "180 days",
    tokens_option_1year: "1 year",
    tokens_action_cancel: "Cancel",
    tokens_action_create: "Create Token",
    tokens_copied: "Token copied to clipboard",
    tokens_copy_failed: "Failed to copy token",

    tokens_delete_title: "Delete Token",
    tokens_delete_message: "Are you sure you want to delete this token? This action cannot be undone and any services using this token will lose access.",
    tokens_delete_cancel: "Cancel",
    tokens_delete_confirm: "Delete",

    // Settings Tab
    settings_title: "Application Settings",
    settings_description: "Configure your application settings here",
    settings_label_name: "Application Name",
    settings_label_description: "Description",

    // Development Notice
    tokens_development_notice: "🚧 Feature under development - Tokens will be used for automated deployments (similar to GitHub Actions)",
  },
}

export type Dictionary = typeof dict

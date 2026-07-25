export type Role = 'user' | 'admin'

export type AuthUser = {
  id: string
  login: string
  role: Role
}

export type ModuleAccess = {
  id: string
  slug: string
  title: string
  description: string
  unlocked: boolean
  completed_count: number
  total_count: number
}

export type SprintAccess = {
  id: string
  slug: string
  title: string
  description: string
  unlocked: boolean
  modules: ModuleAccess[]
  started_at?: string
  deadline_at?: string
  completed_at?: string
  is_overdue: boolean
  show_timer: boolean
}

export type HomeView = {
  feedback_required: boolean
  telegram_bot_username?: string
  sprints: SprintAccess[]
}

export type SubmoduleAccess = {
  id: string
  slug: string
  title: string
  unlocked: boolean
  completed: boolean
}

export type ModuleView = {
  module: ModuleAccess
  submodules: SubmoduleAccess[]
}

export type MenuItem = {
  type: 'sprint' | 'module' | 'submodule'
  id: string
  slug: string
  title: string
  unlocked: boolean
  completed?: boolean
  children?: MenuItem[]
}

export type SubmoduleView = {
  submodule: SubmoduleAccess
  module: ModuleAccess
  body_md: string
  menu: MenuItem[]
}

export type CreatedUser = {
  id: string
  login: string
  password: string
  role: Role
}

export type AdminSubmodule = {
  id: string
  module_id: string
  slug: string
  title: string
  position: number
  is_published: boolean
}

export type AdminModule = {
  id: string
  sprint_id: string
  slug: string
  title: string
  description: string
  position: number
  is_published: boolean
  submodules: AdminSubmodule[]
}

export type AdminSprint = {
  id: string
  slug: string
  title: string
  description: string
  position: number
  is_published: boolean
  modules: AdminModule[]
}

export type AdminTree = {
  sprints: AdminSprint[]
}

export type AdminUser = {
  id: string
  login: string
  role: Role
  is_blocked: boolean
  telegram_linked: boolean
  created_at: string
}

export type FeedbackItem = {
  id: string
  status: string
  requested_at: string
  answered_at?: string
  answer_text: string
}

export type FeedbackSchedule = {
  next_round_at: string
  last_round_at?: string
  seconds_to_next: number
  message_text: string
}

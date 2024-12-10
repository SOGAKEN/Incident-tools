// レスポンス（対応）の型
interface Response {
    ID: number
    IncidentID: number
    Datetime: string // ISO 8601形式の日時文字列
    Status: string
    Vender: number
    Responder: string
    Content: string
}

// インシデントの型
interface Incident {
    ID: number
    Datetime: string // ISO 8601形式の日時文字列
    Status: Statuces
    Vender: number
    Assignee: string
    MessageID: string
    APIData: GAIResponse
    Responses: Response[]
    Relations: any[] // 現在は空配列ですが、必要に応じて型を定義可能
}

interface IncidentDailog {
    Data: Data
}

interface Data {
    EmailData: EmailData
    Incident: Incident
}

interface Statuces {
    ID: number
    Name: string
    Code: number
}

interface EmailData {
    ID: number
    FROM: string
    subject: string
    Incident: Incident
    from: string
    CreatedAt: number
    body: string
    Relations: any[]
    message_id: string
}

interface GAIResponse {
    ID: string
    IncidentID: string
    TaskID: string
    WorkflowRunID: string
    WorkflowID: string
    Status: Statuces
    Body: string
    User: string
    Host: string
    Priority: string
    Judgment: string
    Subject: string
    From: string
    Place: string
    IncidentText: string
    Time: string
    Final: string
    ElapsedTime: number
    TotalTokens: number
    TotalSteps: number
    CreatedAt: number
    FinishedAt: number
    Error: string
    UpdatedAt: string
    WorkflowLogs: string
    Sender: string
}

interface StatusCount {
    count: number
    status: string
}

interface Meta {
    total: number
    pages: number
    page: number
    limit: number
}

// APIレスポンス全体の型
interface IncidentResponse {
    data: EmailData[]
    meta: Meta
    status_counts: StatusCount[]
    unique_assignees: string[]
}

// Status と Priority の型を定数として定義（オプション）
export const IncidentStatus = {
    UNTOUCHED: '未着手',
    IN_PROGRESS: '対応中',
    COMPLETED: '完了'
} as const

export const IncidentPriority = {
    HIGH: '高',
    MEDIUM: '中',
    LOW: '低'
} as const

// 上記定数からユニオン型を生成（オプション）
export type IncidentStatusType = (typeof IncidentStatus)[keyof typeof IncidentStatus]
export type IncidentPriorityType = (typeof IncidentPriority)[keyof typeof IncidentPriority]

// すべての型をエクスポート
export type { Response as IncidentResponse, Incident, IncidentResponse as IncidentsApiResponse, IncidentDailog, Data, EmailData }

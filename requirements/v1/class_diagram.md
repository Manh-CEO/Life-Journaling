@startuml
skinparam packageStyle rectangle
skinparam class {
    BackgroundColor White
    ArrowColor #2c3e50
    BorderColor #2c3e50
}

title Clean Architecture: Go Backend Class Diagram

package "1. Domain Layer (Entities)" #f9f9f9 {
    class User {
        + ID: UUID
        + Email: string
        + Timezone: string
        + AnchorDate: time.Time
        + PromptHour: int
    }

    class Memory {
        + ID: UUID
        + UserID: UUID
        + EntryDate: time.Time
        + Location: string
        + Content: string
        + Sentiment: string
        + IsManualEntry: bool
    }

    class EngagementLog {
        + ID: UUID
        + UserID: UUID
        + RawEmailText: string
        + Status: string
    }
}

package "2. Use Case Layer (Interfaces & Logic)" #e8f4f8 {
    
    interface IMemoryRepository {
        + Create(memory: Memory): error
        + GetByUserID(userID: UUID): []Memory
        + SearchSemantic(query: string): []Memory
    }

    interface IUserRepository {
        + GetUsersForPrompt(hour: int, day: int): []User
    }

    interface IEmailProvider {
        + SendPrompt(email: string, template: string): error
    }

    interface ILLMProvider {
        + ExtractMemoryData(text: string): Memory, error
    }

    class MemoryService <<Use Case>> {
        - repo: IMemoryRepository
        + GetTimeline(userID: UUID): []Memory
        + AddManualEntry(userID: UUID, content: string): error
    }

    class EngagementService <<Use Case>> {
        - userRepo: IUserRepository
        - emailProvider: IEmailProvider
        + TriggerHourlyPrompts(currentHour: int): error
        + TriggerAnnualPortraits(currentDate: time.Time): error
    }

    class IngestionService <<Use Case>> {
        - memoryRepo: IMemoryRepository
        - llmProvider: ILLMProvider
        + ProcessInboundWebhook(rawEmail: string, senderEmail: string): error
    }
}

package "3. Delivery Layer (Handlers/Controllers)" #e8fae8 {
    class RESTGatewayHandler {
        - memoryUC: MemoryService
        + HandleGetTimeline(ctx: Context)
        + HandlePostMemory(ctx: Context)
    }

    class CronHandler {
        - engagementUC: EngagementService
        + HandleHourlyTrigger(ctx: Context)
    }

    class WebhookHandler {
        - ingestionUC: IngestionService
        + HandleCloudflareInbound(ctx: Context)
    }
}

package "4. Infrastructure Layer (Adapters)" #fceced {
    class SupabaseRepository {
        - db: sql.DB
        + Create(memory: Memory): error
        + GetByUserID(userID: UUID): []Memory
        + GetUsersForPrompt(hour: int, day: int): []User
    }

    class ResendAdapter {
        - client: ResendClient
        + SendPrompt(email: string, template: string): error
    }

    class GeminiLLMAdapter {
        - client: GeminiClient
        + ExtractMemoryData(text: string): Memory, error
    }
}

' Relationships: Delivery depends on Use Cases
RESTGatewayHandler --> MemoryService : calls
CronHandler --> EngagementService : calls
WebhookHandler --> IngestionService : calls

' Relationships: Use Cases depend on Domain Entities
MemoryService ..> Memory : uses
EngagementService ..> User : uses
IngestionService ..> Memory : uses
IngestionService ..> EngagementLog : uses

' Relationships: Use Cases define Interfaces (Dependency Inversion)
MemoryService --> IMemoryRepository
EngagementService --> IUserRepository
EngagementService --> IEmailProvider
IngestionService --> ILLMProvider
IngestionService --> IMemoryRepository

' Relationships: Infrastructure implements Use Case Interfaces
SupabaseRepository ..|> IMemoryRepository
SupabaseRepository ..|> IUserRepository
ResendAdapter ..|> IEmailProvider
GeminiLLMAdapter ..|> ILLMProvider

@enduml
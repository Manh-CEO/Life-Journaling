### 1. High-Level Architecture Overview

We are utilizing a purely serverless, "scale-to-zero" architecture. Instead of running expensive servers 24/7, our compute and databases spin up dynamically on demand and power down completely when idle.

* **Client Layer:** Mobile App (iOS/Android) communicating with Edge-deployed APIs.
* **Ingestion & Engagement Layer:** Serverless message queues, free-tier email providers, and edge-based email routing.
* **Intelligence Layer (AI Workers):** Edge functions handling asynchronous LLM tasks (parsing, sentiment, embeddings) via generous developer APIs.
* **Data Storage Layer:** Serverless PostgreSQL for core relational data, Serverless Vector database for semantic search, and zero-egress Object Storage for media.

---

### 2. Component Design & Epic Mapping

#### A. The "Push/Pull" Engagement Engine (Epics 1 & 5)

To handle user engagement without paying for constant background processes, we rely on event-driven serverless triggers.

* **Distributed Scheduler (Upstash QStash):** A serverless message queue that reliably triggers HTTP endpoints on a schedule (US9, US10). It handles the cron logic for different time zones with zero base cost.
* **Notification Service (Resend):** A lightweight email provider triggered by QStash to send out the weekly prompts and annual photo requests (US1, US2, US13).
* **Inbound Webhook Receiver (Cloudflare Email Routing):** When a user replies to an email, Cloudflare natively catches the inbound message for free and forwards the raw payload to our Edge Worker, instantly queuing it for asynchronous processing.

#### B. Data Intelligence Pipeline (Epic 2)

All heavy AI lifting is done completely in the background to keep the system responsive and costs low.

* **AI Edge Workers (Cloudflare Workers):** These serverless functions wake up to consume the raw emails from the queue.
* **Event Extraction & Sentiment (US3, US4):** The worker strips email signatures and sends the text to the **Google Gemini API** (using its generous free tier). It requests structured JSON output: `{ "date": "...", "location": "...", "events": [...], "sentiment": "joy/nostalgia/sadness" }`.
* **Embedding Generation:** To power the semantic search (US6), the text is converted into mathematical vectors using an embedding model and instantly stored in a Serverless Vector DB.

#### C. Core App API & User Experience (Epics 3 & 4)

When the user opens the app, they hit APIs deployed globally at the "edge," ensuring low latency regardless of where they are in the world.

* **Edge API Gateway (Cloudflare Workers):** Routes all mobile requests. Because they run on the edge (close to the user), they are incredibly fast and cost fractions of a cent per million requests.
* **Profile & Timeline Service:** Fetches the user's dashboard data (US15) and timeline (US5) directly from the serverless database.
* **Semantic Search Service (US6):** When a user searches "When was the last time I felt happy?", the Worker converts the query into a vector and queries **Pinecone Serverless** for nearest-neighbor matches, returning the relevant journal entries.
* **Manual Override Service:** Standard CRUD operations hitting our database via REST, allowing users to log, edit, or delete entries (US11, US12).

#### D. Media & Storage (US2, US14)

* **Blob Storage (Cloudflare R2 or Supabase Storage):** Stores the annual portraits. R2 is particularly cost-effective here because it charges **$0 for egress (bandwidth)**, which is usually the most expensive part of serving images to a mobile app timeline (US15).

#### E. Security & Legacy (Epic 6)

* **Encryption (US7):** We rely on **Supabase Row Level Security (RLS)** to ensure users can only ever access their own rows. For text encryption, we use application-level envelope encryption inside the Cloudflare Worker, securely storing the master keys in Cloudflare Secrets.
* **Export Worker (US8):** PDF generation is shifted to a background Cloudflare Worker (or an edge function with a slightly higher timeout limit) that generates the document, saves it to R2, and triggers a push notification to the user upon completion.

---

### 3. Database Strategy

We are using a polyglot approach, strictly selecting platforms that offer scale-to-zero capabilities and generous free tiers for bootstrapping:

| Database Type | Technology Choice | Purpose |
| --- | --- | --- |
| **Relational DB** | **Supabase** (Serverless Postgres) | User profiles, authentication, timeline metadata, and encrypted text. Scales to zero when inactive. |
| **Vector DB** | **Pinecone Serverless** | Storing and querying vector embeddings for Natural Language Search (US6). Completely usage-based billing. |
| **Message Queue** | **Upstash QStash** | Managing the weekly email triggers and decoupling the inbound email parsing from the UI. |
| **Blob Storage** | **Cloudflare R2** | Storing original and thumbnail versions of annual portraits with zero bandwidth costs. |

---

### 4. Critical Bottlenecks to Watch Out For

While this architecture drops your baseline cost to $0, it introduces new constraints you must manage:

1. **Cold Starts:** Because the infrastructure "sleeps" when no one is using it, the very first user to open the app after a long period of inactivity might experience a 1-2 second delay while the Supabase instance or Cloudflare Worker spins up.
2. **Third-Party Rate Limits:** Utilizing free tiers for the Google Gemini API and Resend means you are bound by strict requests-per-minute limits. Your QStash queue must be perfectly configured to "drip feed" the AI processing so you don't get throttled during a spike in email replies.
3. **Database Size Limits:** Supabase's free tier caps your database size at 500MB. If you go viral, you will hit this limit quickly and the database will switch to read-only mode until you upgrade to a paid tier.

```planuml
@startuml
!include https://raw.githubusercontent.com/plantuml-stdlib/C4-PlantUML/master/C4_Container.puml

LAYOUT_TOP_DOWN()
LAYOUT_WITH_LEGEND()

title C4 Level 2: Container Diagram - Personal Memory App (Abstracted)

Person(user, "App User", "Records their life events via mobile app and email replies.")

System_Ext(email_sys, "External Email Service", "Handles high-volume outbound prompts and routes inbound user replies.")
System_Ext(llm_sys, "External Intelligence Service", "Large Language Model API for parsing text, extracting metadata, and tagging sentiment.")
System_Ext(kms_sys, "Key Management Service", "External or managed service for securely storing encryption keys.")

System_Boundary(memory_system, "Personal Memory Backend System") {
    
    Container(mobile_app, "Mobile Client", "Native App", "Provides the user interface for dashboards, timelines, and manual overrides.")
    Container(api_gateway, "API Gateway", "Gateway Component", "Acts as the single entry point, handling auth, rate limiting, and routing.")
    
    Container_Boundary(ingestion_layer, "Ingestion & Engagement Boundary") {
        Container(scheduler, "Distributed Scheduler", "Job Orchestrator", "Triggers engagement events based on user preferences and timezones.")
        Container(webhook, "Webhook Receiver", "Lightweight Web Server", "Fast-responding endpoint for receiving external events.")
        ContainerQueue(message_broker, "Event Broker", "Message Queue", "Buffers incoming payloads and decouples ingestion from heavy processing.")
    }
    
    Container_Boundary(core_layer, "Core Services Boundary") {
        Container(core_api, "Core Application API", "Web Service", "Handles synchronous requests for profiles, timelines, search, and manual entries.")
    }
    
    Container_Boundary(worker_layer, "Asynchronous Intelligence Boundary") {
        Container(ai_worker, "Intelligence Extraction Worker", "Background Processor", "Parses raw text, maps structured events, and generates vector embeddings.")
        Container(export_worker, "Legacy Export Worker", "Background Processor", "Compiles historical data into downloadable formats.")
    }
    
    Container_Boundary(data_layer, "Data Persistence Boundary") {
        ContainerDb(cache, "Distributed Cache", "In-Memory Store", "Caches frequently accessed dashboard and timeline data for low latency.")
        ContainerDb(db, "Primary Database", "Relational Database", "Stores encrypted user entries, relationships, and application state.")
        ContainerDb(vector_db, "Semantic Search Database", "Vector Store", "Stores mathematical embeddings to power natural language queries.")
        ContainerDb(blob_store, "Object Storage", "Blob Store", "Stores unstructured media like portrait photos and generated documents.")
    }
}

' External & Edge Interactions
Rel(user, mobile_app, "Interacts with UI components", "HTTPS")
Rel(user, email_sys, "Receives prompts & replies via email")
Rel(mobile_app, blob_store, "Uploads media directly using signed URLs", "HTTPS")
Rel(mobile_app, api_gateway, "Makes synchronous data requests", "JSON/HTTPS")

' API Routing
Rel(api_gateway, core_api, "Routes client requests", "Internal RPC/HTTP")

' The Push/Pull Engagement Loop
Rel(scheduler, email_sys, "Triggers scheduled communication logic", "API")
Rel(email_sys, webhook, "Pushes user replies via webhook", "HTTPS")
Rel(webhook, message_broker, "Publishes raw payloads instantly", "TCP")

' Core API Data Reads/Writes
Rel(core_api, cache, "Reads/Writes fast-access data", "TCP")
Rel(core_api, db, "Reads/Writes primary application records", "TCP")
Rel(core_api, vector_db, "Queries semantic matches", "TCP")
Rel(core_api, message_broker, "Publishes manual background jobs", "TCP")

' Asynchronous Workers
Rel(message_broker, ai_worker, "Consumes raw data safely at its own pace", "TCP")
Rel(message_broker, export_worker, "Consumes heavy processing jobs", "TCP")

' AI Worker logic
Rel(ai_worker, llm_sys, "Requests text parsing & sentiment analysis", "HTTPS")
Rel(ai_worker, db, "Persists structured event data", "TCP")
Rel(ai_worker, vector_db, "Persists semantic embeddings", "TCP")

' Export Worker logic
Rel(export_worker, db, "Reads full historical ledger", "TCP")
Rel(export_worker, blob_store, "Saves generated documents", "HTTPS")

' Security
Rel(db, kms_sys, "Encrypts/Decrypts sensitive text via Envelope Encryption", "HTTPS")

@enduml
```

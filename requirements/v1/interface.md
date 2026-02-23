
# Mobile Interface Specification (Flutter V1)

## 1. Onboarding & Configuration Flow

This flow captures the essential user parameters required for the Upstash QStash cron triggers to function correctly.

### 1.1. Authentication Screen

* **Sign Up / Log In:** Utilizes Supabase Auth for standard email authentication.
* **Action:** On successful authentication, creates a record in the `users` table.

### 1.2. Context Setup Screen

* **Timezone Selector:** A dropdown list to capture the user's local timezone for accurate cron triggers.
* **Anchor Date Picker:** A calendar input for the user to select their birthday or chosen milestone date.

### 1.3. Schedule Preferences Screen

* **Prompt Day Selector:** A visual picker allowing the user to select a day of the week from Sunday (0) to Saturday (6).
* **Prompt Hour Selector:** A time picker to select an hour between 0 and 23.
* **Action:** Saves preferences to the `users` table and completes onboarding.

---

## 2. Home Dashboard & Timeline

This is the primary "Home" view. It is designed to create an immediate emotional connection and provide a frictionless way to review past entries.

### 2.1. Personal Identity Header

* **Profile Display:** Shows the user's profile information and current age or milestone.
* **Annual Portrait:** Displays the most recent photo fetched via the `storage_path` from the `portraits` table.

### 2.2. Visual Timeline

* **Structure:** An interactive, vertically scrolling list of memory cards.
* **Interaction:** Tapping a memory card opens a full-screen view with options to edit or delete the entry.

| Memory Card UI Element | Database Mapping (`memories` table) | Display Logic |
| --- | --- | --- |
| **Header** | `entry_date` | Formatted as a readable date. |
| **Location Badge** | `location` | Displayed if the AI extracted a location. |
| **Body Text** | `content` | Shows a preview snippet of the decrypted text. |
| **Mood Tag** | `sentiment` | Visual tag (e.g., color-coded label) for 'joy', 'nostalgia', etc.. |
| **Source Indicator** | `is_manual_entry` | A small icon if the entry was added manually via the app. |

---

## 3. Search & Discovery

This screen leverages the Pinecone Serverless vector database to allow users to find memories using everyday language.

### 3.1. Natural Language Search

* **Search Bar:** A prominent text input field at the top of the screen.
* **Placeholder Text:** "e.g., When was the last time I felt happy?".
* **Results View:** Re-renders the Visual Timeline format using the semantic matches returned by the backend.

---

## 4. Manual Overrides & Entry

These interfaces empower the user to manage their data directly, bypassing the automated email ingestion system.

### 4.1. Add / Edit Memory Modal

* **Date & Time Picker:** Defaults to the current timestamp but allows historical backdating.
* **Location Input:** Optional text field for the location.
* **Text Area:** A large input field for the journal entry content.
* **Save Action:** Inserts or updates the row in the `memories` table with `is_manual_entry` set to TRUE.

### 4.2. Upload Portrait Modal

* **Image Picker:** Invokes the native iOS/Android gallery picker.
* **Year Selector:** An integer input to define the `portrait_year` for timeline placement.
* **Upload Action:** Uploads the blob to Cloudflare R2 / Supabase Storage and saves the path with `is_manual_upload` set to TRUE.

---

## 5. Settings & Legacy

A secondary screen for profile management and data portability.

### 5.1. Legacy Export

* **Export Action:** A primary button labeled "Generate Life Story Book".
* **Logic:** Triggers the background worker to compile the historical data into a PDF and pushes a notification upon completion.

---
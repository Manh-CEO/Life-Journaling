### **V1 Implementation Roadmap (4 Phases)**

#### **Phase 1: The "Brains" (Supabase Setup)**

Instead of managing a complex backend, we use Supabase for your database and authentication.

* **Action:** Deploy the SQL schema provided in the previous step.
* **Why:** It handles user sign-ups, stores their timezones, and keeps a log of every email sent.
* **Cost:** **$0** (Supabase Free Tier).

#### **Phase 2: The "Alarm Clock" (Upstash QStash)**

You need a way to trigger the "Check-in" emails without a server running 24/7.

* **Action:** Create an Upstash account and set up a **Scheduled Task (Cron)**.
* **Schedule:** Set it to hit your Cloudflare Worker URL every hour: `0 * * * *`.
* **Task:** The payload tells the Worker: *"Check who needs a prompt for this hour."*
* **Cost:** **$0** (Upstash gives you 500 free messages per day).

#### **Phase 3: The "Messenger" (Cloudflare Worker + Resend)**

This is the only piece of code you actually need to write. It acts as the bridge.

* **Action:** Deploy a Cloudflare Worker that:
1. Receives the "Hourly" ping from QStash.
2. Queries Supabase: `SELECT email FROM users WHERE local_hour = current_hour`.
3. Loops through the users and calls the **Resend API** to send the email.


* **Code Tip:** Use the `@supabase/supabase-js` and `resend` libraries.
* **Cost:** **$0** (Cloudflare Workers: 100k requests/day; Resend: 3,000 emails/month).

#### **Phase 4: The "Ear" (Cloudflare Email Workers)**

This is the most critical part for a "reply-to-log" system.

* **Action:** Enable **Email Routing** in Cloudflare.
* **Logic:** Set up a rule: *"Any email sent to `reply@yourdomain.com` → Send to Email Worker."*
* **Processing:** The Email Worker receives the raw email, parses the "From" address to identify the user, and saves the text directly into the `engagement_logs` table in Supabase.
* **Cost:** **$0** (Included in Cloudflare's free tier).

---

### **Revised V1 Simplified Document**

| Feature | Phase | Technology | Logic |
| --- | --- | --- | --- |
| **User Onboarding** | Phase 1 | Supabase Auth | User sets timezone and birthday (Anchor Date). |
| **Weekly Prompt** | Phase 2/3 | QStash + Resend | Scheduled trigger sends an email question every week. |
| **Annual Portrait** | Phase 2/3 | QStash + Resend | Sends a special email on the user's "Anchor Date." |
| **Reply Capture** | Phase 4 | CF Email Workers | User replies to the email; text is saved to DB automatically. |

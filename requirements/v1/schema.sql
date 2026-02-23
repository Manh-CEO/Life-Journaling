-- ====================================================================================
-- PROJECT: Personal Memory App (V1)
-- DB: Supabase (PostgreSQL)
-- ====================================================================================

-- 1. USERS TABLE (Epics 1 & 5: Context, Scheduling, Anchor Dates)
-- Links to Supabase's native auth.users table.
CREATE TABLE public.users (
    id UUID REFERENCES auth.users(id) ON DELETE CASCADE PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    timezone TEXT DEFAULT 'UTC', -- US9: For accurate Cron triggers
    anchor_date DATE, -- US2 & US9: Birthday or chosen milestone date
    prompt_day_of_week INT CHECK (prompt_day_of_week >= 0 AND prompt_day_of_week <= 6), -- US10: 0=Sun, 6=Sat
    prompt_hour INT CHECK (prompt_hour >= 0 AND prompt_hour <= 23), -- Matches local_hour logic
    created_at TIMESTAMP WITH TIME ZONE DEFAULT timezone('utc'::text, now()) NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT timezone('utc'::text, now()) NOT NULL
);

-- 2. ENGAGEMENT LOGS (Phase 4 / Epic 1: The "Ear")
-- Cloudflare Email Worker dumps the raw inbound email text here instantly.
CREATE TABLE public.engagement_logs (
    id UUID DEFAULT extensions.uuid_generate_v4() PRIMARY KEY,
    user_id UUID REFERENCES public.users(id) ON DELETE CASCADE NOT NULL,
    raw_email_text TEXT NOT NULL, -- The raw reply before AI processing
    status TEXT DEFAULT 'pending' CHECK (status IN ('pending', 'processed', 'failed')),
    received_at TIMESTAMP WITH TIME ZONE DEFAULT timezone('utc'::text, now()) NOT NULL
);

-- 3. MEMORIES / JOURNAL ENTRIES (Epic 2, 3, & 4: Data Intelligence & Overrides)
-- AI Workers populate this after parsing engagement_logs, or users add manually.
CREATE TABLE public.memories (
    id UUID DEFAULT extensions.uuid_generate_v4() PRIMARY KEY,
    user_id UUID REFERENCES public.users(id) ON DELETE CASCADE NOT NULL,
    entry_date TIMESTAMP WITH TIME ZONE NOT NULL, -- US3: Extracted by AI or manually set
    location TEXT, -- US3: Extracted by AI
    content TEXT NOT NULL, -- US7: Encrypted text (managed at application/worker level)
    sentiment TEXT, -- US4: AI tagged mood (e.g., 'joy', 'nostalgia')
    is_manual_entry BOOLEAN DEFAULT FALSE, -- US11: True if entered via UI
    created_at TIMESTAMP WITH TIME ZONE DEFAULT timezone('utc'::text, now()) NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT timezone('utc'::text, now()) NOT NULL
);

-- 4. PORTRAITS / MEDIA (Epic 1 & 4: Annual Portraits)
-- Stores metadata for the images saved in Cloudflare R2 / Supabase Storage.
CREATE TABLE public.portraits (
    id UUID DEFAULT extensions.uuid_generate_v4() PRIMARY KEY,
    user_id UUID REFERENCES public.users(id) ON DELETE CASCADE NOT NULL,
    storage_path TEXT NOT NULL, -- Pointer to the R2/Blob object
    portrait_year INT NOT NULL, -- US15: To display on the timeline easily
    is_manual_upload BOOLEAN DEFAULT FALSE, -- US14: True if uploaded from gallery
    captured_at TIMESTAMP WITH TIME ZONE DEFAULT timezone('utc'::text, now()) NOT NULL
);

-- ====================================================================================
-- ROW LEVEL SECURITY (RLS) - Epic 6 (US7: Privacy & Encryption)
-- ====================================================================================

-- Enable RLS on all tables
ALTER TABLE public.users ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.engagement_logs ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.memories ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.portraits ENABLE ROW LEVEL SECURITY;

-- USERS POLICIES
CREATE POLICY "Users can view own profile" ON public.users FOR SELECT USING (auth.uid() = id);
CREATE POLICY "Users can update own profile" ON public.users FOR UPDATE USING (auth.uid() = id);

-- ENGAGEMENT LOGS POLICIES
CREATE POLICY "Users can view own logs" ON public.engagement_logs FOR SELECT USING (auth.uid() = user_id);
-- Note: Insert policy for engagement_logs might need to allow the Cloudflare Worker role to insert

-- MEMORIES POLICIES (US11, US12: Manual control & Management)
CREATE POLICY "Users can view own memories" ON public.memories FOR SELECT USING (auth.uid() = user_id);
CREATE POLICY "Users can insert own memories" ON public.memories FOR INSERT WITH CHECK (auth.uid() = user_id);
CREATE POLICY "Users can update own memories" ON public.memories FOR UPDATE USING (auth.uid() = user_id);
CREATE POLICY "Users can delete own memories" ON public.memories FOR DELETE USING (auth.uid() = user_id);

-- PORTRAITS POLICIES
CREATE POLICY "Users can view own portraits" ON public.portraits FOR SELECT USING (auth.uid() = user_id);
CREATE POLICY "Users can insert own portraits" ON public.portraits FOR INSERT WITH CHECK (auth.uid() = user_id);
CREATE POLICY "Users can delete own portraits" ON public.portraits FOR DELETE USING (auth.uid() = user_id);

-- ====================================================================================
-- TRIGGERS & FUNCTIONS
-- ====================================================================================

-- Function to automatically update 'updated_at' columns
CREATE OR REPLACE FUNCTION update_modified_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply triggers
CREATE TRIGGER update_users_modtime BEFORE UPDATE ON public.users FOR EACH ROW EXECUTE PROCEDURE update_modified_column();
CREATE TRIGGER update_memories_modtime BEFORE UPDATE ON public.memories FOR EACH ROW EXECUTE PROCEDURE update_modified_column();

-- Function to automatically create a user row when a new Supabase Auth user signs up
CREATE OR REPLACE FUNCTION public.handle_new_user()
RETURNS TRIGGER AS $$
BEGIN
  INSERT INTO public.users (id, email)
  VALUES (NEW.id, NEW.email);
  RETURN NEW;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Trigger to fire on user signup
CREATE TRIGGER on_auth_user_created
  AFTER INSERT ON auth.users
  FOR EACH ROW EXECUTE PROCEDURE public.handle_new_user();
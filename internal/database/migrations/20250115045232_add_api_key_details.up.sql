ALTER TABLE public.keys_metadata ADD COLUMN api_key_hash text;
ALTER TABLE public.keys_metadata ADD COLUMN locked boolean DEFAULT false;

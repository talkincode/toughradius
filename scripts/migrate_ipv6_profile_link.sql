-- Migration: Add ProfileLinkMode and rename IPv6Prefix to IPv6PrefixPool
-- Date: 2025-11-25
-- Description: Support dynamic/static profile linking and clarify IPv6 pool semantics

-- Step 1: Rename IPv6Prefix to IPv6PrefixPool in radius_profile table
ALTER TABLE radius_profile RENAME COLUMN ipv6_prefix TO ipv6_prefix_pool;

-- Step 2: Add new fields to radius_user table
ALTER TABLE radius_user ADD COLUMN IF NOT EXISTS ipv6_prefix_pool VARCHAR(100) DEFAULT '';
ALTER TABLE radius_user ADD COLUMN IF NOT EXISTS profile_link_mode INTEGER DEFAULT 0;

-- Step 3: Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_radius_user_profile_link_mode ON radius_user(profile_link_mode);

-- Step 4: Update comments (PostgreSQL specific)
COMMENT ON COLUMN radius_profile.ipv6_prefix_pool IS 'IPv6 prefix pool name for NAS-side allocation (e.g., "pool-vip")';
COMMENT ON COLUMN radius_user.ipv6_prefix_pool IS 'IPv6 prefix pool name, inherited from profile or user-specific override';
COMMENT ON COLUMN radius_user.profile_link_mode IS 'Profile link mode: 0=static (snapshot on creation), 1=dynamic (real-time from profile)';

-- Note: For SQLite, comments are not supported. The schema will be auto-migrated by GORM.

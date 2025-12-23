---
selectors:
  database: postgres
  feature: auth
---

# Database Setup Instructions

This command provides database setup instructions. When this command is used in a task,
it will automatically include rules that are tagged with `database: postgres` and 
`feature: auth` in their frontmatter.

## PostgreSQL Configuration

Connect to PostgreSQL:
```bash
psql -U ${db_user} -d ${db_name}
```

## Authentication Setup

Configure authentication tables and initial data.

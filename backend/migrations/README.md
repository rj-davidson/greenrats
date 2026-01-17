# Migrations

This directory contains Atlas versioned migrations for the database schema.

## Generating Migrations

After modifying `ent/schema/*.go` files:

```bash
# 1. Regenerate Ent code
go generate ./ent

# 2. Generate migration (requires Docker running)
atlas migrate diff <migration_name> --env local

# 3. Review the generated SQL in this directory

# 4. Apply migration
atlas migrate apply --env local
```

## Initial Migration

To generate the initial schema migration:

```bash
atlas migrate diff initial_schema --env local
```

## Checking Status

```bash
atlas migrate status --env local
```

## After Manual Edits

If you manually edit a migration file, rehash:

```bash
atlas migrate hash --env local
```

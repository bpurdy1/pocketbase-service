# Collections Schema

Multi-tenant data model built on PocketBase. All collections are created programmatically on server startup via `OnServe()` hooks.

## Entity Relationship

```
_superusers (built-in)
  Bypasses ALL access rules. Platform-level administration.

users (built-in, extended)
  ├── phone          text      (required, E.164 format)
  ├── role           select    (user | agent | admin)
  │
  ├──< org_members   join table
  │     ├── organization  relation  → organizations
  │     └── role          select    (owner | admin | member)
  │
  └──< settings      1:1 per user
        ├── email_notifications  bool
        ├── sms_notifications    bool
        ├── theme                select  (light | dark | system)
        ├── timezone             text
        └── preferences          json

organizations
  ├── name       text   (required)
  ├── slug       text   (required, unique)
  ├── website    url
  ├── phone      text
  ├── address    text
  ├── city       text
  ├── state      text   (2 char)
  ├── zip_code   text
  │
  └──< properties
        ├── property_name    text    (required)
        ├── address          text    (required)
        ├── city             text    (required)
        ├── state            text    (2 char)
        ├── zip_code         text    (xxxxx or xxxxx-xxxx)
        ├── county           text
        ├── year_built       number
        ├── number_of_units  number
        ├── building_sf      number
        └── lot_sf           number
```

## Roles

### Platform Roles (`users.role`)

| Role    | Description                        |
|---------|------------------------------------|
| `user`  | Default role assigned on signup    |
| `agent` | Real estate agent                  |
| `admin` | Platform administrator             |

### Organization Roles (`org_members.role`)

| Role     | Description                                  |
|----------|----------------------------------------------|
| `owner`  | Full control. Auto-assigned to org creator   |
| `admin`  | Can manage members and org data              |
| `member` | Read-only access to org data                 |

## Access Rules

### organizations

| Action | Rule                          |
|--------|-------------------------------|
| List   | User is a member of the org   |
| View   | User is a member of the org   |
| Create | Any authenticated user        |
| Update | Org owner or admin            |
| Delete | Org owner only                |

### org_members

| Action | Rule                          |
|--------|-------------------------------|
| List   | User is a member of the org   |
| View   | User is a member of the org   |
| Create | Org owner or admin            |
| Update | Org owner or admin            |
| Delete | Org owner or admin            |

Unique constraint: one membership per user per organization.

### properties

| Action | Rule                          |
|--------|-------------------------------|
| List   | User is a member of the org   |
| View   | User is a member of the org   |
| Create | Org owner or admin            |
| Update | Org owner or admin            |
| Delete | Org owner or admin            |

### settings

| Action | Rule                                    |
|--------|-----------------------------------------|
| List   | Owner or platform admin                 |
| View   | Owner or platform admin                 |
| Create | Owner or platform admin                 |
| Update | Owner or platform admin                 |
| Delete | Owner or platform admin                 |

Unique constraint: one settings record per user.

### _superusers

PocketBase built-in. Superusers bypass all collection access rules and have full read/write access to everything. Managed via the `PB_ADMIN_EMAIL` and `PB_ADMIN_PASS` environment variables.

## Hooks

| Event                              | Action                                     |
|------------------------------------|--------------------------------------------|
| User signup                        | Auto-assigns `role = "user"`               |
| Organization created               | Auto-creates `org_members` with `"owner"`  |

## File Structure

```
internal/collections/
  roles.go                          Role constants and enums
  ensure.go                         Users collection (phone, role fields)
  orginization.go                   Organizations collection + owner hook
  org_members.go                    Join table (user + org + role)
  settings.go                       Per-user settings
  realestate/
    property-collection.go          Properties (org-scoped)
```

## API Examples

### Sign up
```
POST /api/collections/users/records
{
  "email": "agent@firm.com",
  "username": "janedoe",
  "password": "SecurePass1!",
  "passwordConfirm": "SecurePass1!",
  "phone": "+15551234567"
}
// role auto-assigned to "user"
```

### Create an organization
```
POST /api/collections/organizations/records
Authorization: Bearer <user_token>
{
  "name": "Acme Realty",
  "slug": "acme-realty",
  "state": "CA"
}
// creator auto-added as org owner
```

### Add a member to the org
```
POST /api/collections/org_members/records
Authorization: Bearer <owner_or_admin_token>
{
  "user": "<user_id>",
  "organization": "<org_id>",
  "role": "member"
}
```

### Create a property
```
POST /api/collections/properties/records
Authorization: Bearer <org_admin_token>
{
  "organization": "<org_id>",
  "property_name": "Sunset Apartments",
  "address": "123 Sunset Blvd",
  "city": "Los Angeles",
  "state": "CA",
  "zip_code": "90028",
  "number_of_units": 12,
  "building_sf": 15000,
  "lot_sf": 8500
}
```

### List properties (auto-filtered to user's org)
```
GET /api/collections/properties/records
Authorization: Bearer <user_token>
```

---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "qernal_registry_secret Data Source - qernal"
subcategory: ""
description: |-
  
---

# qernal_registry_secret (Data Source)





<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) Name of the registry secret
- `project_id` (String)

### Read-Only

- `date` (Attributes) (see [below for nested schema](#nestedatt--date))
- `registry` (String) url of the registry
- `revision` (Number)

<a id="nestedatt--date"></a>
### Nested Schema for `date`

Optional:

- `created_at` (String)
- `updated_at` (String)
---
description: Example of file references in rules
---

# File Reference Feature Example

This rule demonstrates how to use file references in rule files.

## Database Configuration

When working with database code, always refer to the database configuration in @config/database.yaml for connection settings.

## API Endpoints

The API endpoint configuration can be found in @config/api.json. Ensure all API calls use the endpoints defined in this file.

## Best Practices

File references are useful for:
- Including configuration files that should be considered
- Showing example code that illustrates patterns
- Referencing schema definitions
- Including test data or fixtures

Simply use `@filepath` syntax and the file contents will be automatically included as formatted code blocks.

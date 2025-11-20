---
task_name: code-review-with-files
description: Review specific code files
---

# Code Review with File References

This task demonstrates the file reference feature, which allows you to include file contents directly in your task prompts.

## Review These Files

### Main Component
Please review @src/components/Button.tsx for:
- Code quality and maintainability
- Performance optimizations
- Accessibility compliance
- TypeScript best practices

### Configuration
Check the configuration in @config/app.yaml to ensure:
- All required settings are present
- Values are appropriate for the environment
- No sensitive data is exposed

## Instructions

Compare the implementation in the component against the configuration requirements. Suggest any improvements or issues you find.

## Note

File references like `@filepath` are automatically expanded by the coding-context CLI. The actual file contents will be included when this task is executed, formatted as code blocks with syntax highlighting.

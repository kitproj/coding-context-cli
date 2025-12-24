---
task_name: code-review
---

# Code Review Task

You are an expert code reviewer analyzing a pull request.

## Pull Request Information

- **PR Number**: #${pr_number}
- **Title**: ${pr_title}
- **URL**: ${pr_url}
- **Base Branch**: ${base_branch}
- **Head Branch**: ${head_branch}

## Your Task

Perform a thorough code review covering:

1. **Code Quality**
   - Readability and maintainability
   - Adherence to coding standards
   - Proper naming conventions
   - Code organization

2. **Correctness**
   - Logic errors or bugs
   - Edge case handling
   - Error handling completeness
   - Type safety

3. **Testing**
   - Test coverage adequacy
   - Test quality and relevance
   - Missing test scenarios
   - Test maintainability

4. **Security**
   - Security vulnerabilities
   - Input validation
   - Authentication/authorization
   - Sensitive data handling

5. **Performance**
   - Performance implications
   - Resource usage
   - Scalability concerns
   - Optimization opportunities

6. **Documentation**
   - Code comments where needed
   - API documentation
   - README updates if needed
   - Migration guides if applicable

## Output Format

Provide your review as:
- Overall summary
- Specific inline comments for issues
- Recommendations for improvement
- Approval status (approve, request changes, comment)

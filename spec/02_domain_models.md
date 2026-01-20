### 02_domain_models.md

Struct: Project
- Name: string
- Description: string
- Slug: string

Struct: BlogPost
- Title: string
- Slug: string
- Date: string
- Excerpt: string
- Content: template.HTML

Struct: PageData
- Title: string
- MetaDescription: string
- CanonicalURL: string
- SocialImageURL: string
- GitHubUser: string
- Projects: []Project
- BlogPosts: []BlogPost
- CurrentPost: BlogPost
- CurrentProject: Project
- ProfileImage: string
- ReadmeHTML: template.HTML
- ActiveNav: string
- AdminError: string
- AdminAuthenticated: bool
- Messages: []ContactMessage

Struct: ContactMessage
- ID: int
- Email: string
- Message: string
- CreatedAt: string

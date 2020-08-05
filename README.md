# Echo REST API - sample project

This project contains a sample REST API built using Go and the Echo framework,
having a MongoDB database.

It contains configurations from top to bottom to meet almost all needs of a small scale REST API,
including:
1. CSRF cookies
2. JWT
3. Password hashing using Argon2
4. CORS
5. Panic recovery
6. Prometheus metrics
7. Data caching
8. SMTP for email sending (used for user invites)
9. Custom error handling and logging
10. Much more

Notes: even though most of the code present in this project is production ready,
please do not use this project as-is in a production environment!
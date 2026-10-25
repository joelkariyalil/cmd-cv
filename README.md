# CMD-CV (drop-box for txt)

[![Code Deployment](https://github.com/joelkariyalil/cmd-cv/actions/workflows/cd.yml/badge.svg)](https://github.com/joelkariyalil/cmd-cv/actions/workflows/cd.yml) [![code-integration](https://github.com/joelkariyalil/cmd-cv/actions/workflows/ci.yml/badge.svg)](https://github.com/joelkariyalil/cmd-cv/actions/workflows/ci.yml)

**A simple, serverless text vault for sharing messages between devices.**

## The idea

I wanted to copy my `.golangci.yml` configuration from one device to another without manually sending it through WhatsApp or another messaging app. CMD-CV solves this by letting you store text and retrieve it using a unique code.

## Architecture

- **Frontend:** HTML, CSS, JavaScript, hosted on GitHub Pages.
- **Backend:** Go running on AWS Lambda.
- **API:** Amazon API Gateway HTTP API.
- **Storage:** Amazon DynamoDB.
- **Infrastructure:** AWS SAM / CloudFormation.

## Deployment

Prerequisites: AWS CLI, AWS SAM CLI, Go, and an AWS account.

Deploy the backend:

```bash
sam build
sam deploy --guided
```

Configure `config.js` with your API Gateway URL, enable CORS for your GitHub Pages origin, and deploy the frontend through GitHub Pages.

## Security

Keep your secret codes private. Anyone with a valid code may be able to access the associated message. Avoid storing passwords, tokens, or other highly sensitive information.

---

Built to make transferring text between devices a little less annoying. 

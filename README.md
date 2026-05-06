# Project phi-redactor

Performant PHI redaction in Go designed to avoid unnecessary language model use.

## Mental Model

1. lexer emits tokens
2. detectors identify candidates for names or other amiguous things, and findings for finding items that can be identified easily via regex, such as phone numbers, email addresses, etc.). Name detector works via aho-corasick over a list of names from ssa and census bureau, but for now, is only using names that commonly double as verbs.
3. resolver handles conflicts between candidates. For example, deciding if "will" is a name or not. Emits findings after. in the future, this could incorporate a small NER model for truly ambiguous cases, but for now will simply pass if info is too ambiguous.
4. redactor transforms text (changes "Dr. Grace saw Hope today" into Dr. Grace saw [NAME] today" etc.)
5. logger records what happened safely (input and output if app_env is local, otherwise only logs output)

## endpoints

- `POST /redact`
- `GET /healthz`
- `GET /readyz`
- `GET /metrics`

## Getting Started

These instructions will get you a copy of the project up and running on your local machine for development and testing purposes. See deployment for notes on how to deploy the project on a live system.

## MakeFile

Run build make command with tests

```bash
make all
```

Build the application

```bash
make build
```

Run the application

```bash
make run
```

Live reload the application:

```bash
make watch
```

Run the test suite:

```bash
make test
```

Clean up binary from the last build:

```bash
make clean
```

# Project phi-redactor

Performant PHI redaction in Go.

## Mental Model

1. lexer emits tokens
2. detectors identify candidates
3. resolver handles conflicts between candidates (deciding if "will" is a name or not, etc.)
4. redactor transforms text (changes "Dr. Grace saw Hope today" into Dr. Grace saw [NAME] today" etc.)
5. logger records what happened safely (input and output if app_env is local, otherwise only logs output)

Right now, only names are implemented, so resolver isn't doing anything.

## Todo

- [ ] POST /redact
- [ ] GET /healthz
- [ ] GET /readyz
- [ ] GET /metrics
- [ ] safe structured detection logs
- [ ] benchmark suite
- [ ] synthetic PHI examples
- [ ] configurable detectors

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

# Sequencer for SYB

## Setup
```bash
cp .env.example .env
brew install go-task # for running various tasks, especially tests
brew install --cask tableplus # Native GUI tool for relational databases
```

## Running Tests
```bash
task test-<name> # for example: task test-historydb
```
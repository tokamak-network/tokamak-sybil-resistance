# Sequencer for SYB

## Setup
```bash
cp .env.example .env
brew install go-task # for running various tasks, especially tests
```

## Running Tests
```bash
task test-<name> # for example: task test-historydb
```

## Architecture

### E2E flow
<img src="../doc/images/sequencer_e2e_flow.png" />

### Sync flow
<img src="../doc/images/sequencer_sync_flow.png" />

### Coord flow
<img src="../doc/images/sequencer_coord_flow.png" />
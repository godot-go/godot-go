# native-call-pinning Delta

## ADDED Requirements

### Requirement: Memory Exposed Across A Paired Release Callback Is Not Accumulated

Memory the engine holds a pointer to between an exposure callback and its paired release callback (a get/free callback sequence) SHALL remain valid from exposure until the paired release callback returns, SHALL be retained on a mechanism whose lifetime matches that pair or longer, and SHALL NOT be pinned on the program-lifetime pinner per exposure. Repeating the exposure/release sequence N times SHALL NOT grow pinner bookkeeping or heap retention proportionally to N.

#### Scenario: Property list churn leaves no pins behind

- **WHEN** the engine performs thousands of get_property_list/free_property_list cycles against registered classes
- **THEN** the program-lifetime pinner holds no entry attributable to any of the exposed arrays after the cycles complete
- **AND** heap retention does not grow proportionally to the cycle count

#### Scenario: Exposed memory valid until the paired release

- **WHEN** an exposure callback has returned but its paired release callback has not yet run
- **THEN** the release callback can read the exposed memory — iterate its elements and free its contents — with no corruption

# Documentation

## Choose Your Path

- User docs: `docs/user/README.md`
- Install details: `INSTALL.md`
- Threat model: `docs/user/THREAT_MODEL.md`
- Developer docs: `docs/dev/README.md`
- Product concept: `docs/concepts/product-concept.md`
- Security trust model: `docs/concepts/security-trust-model.md`

## Current Structure

- `docs/concepts/`: product definition and positioning
- `docs/dev/spec/`: target implementation contracts
- `docs/user/`: user-facing guidance
- `docs/plans/`: design notes and implementation planning

The repository is in the middle of a redesign from a deny-only guard toward a
declarative permission policy proxy. Current user-facing behavior is documented
in `README.md` and `docs/user/`. Proposed or planned behavior lives under
`docs/dev/spec/` and must not be treated as shipped unless its frontmatter
status is `implemented`.

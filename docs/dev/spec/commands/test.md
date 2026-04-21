---
title: "cmdproxy test"
status: proposed
date: 2026-04-21
---

# cmdproxy test

## Purpose

`cmdproxy test` is deprecated.
Its former role is subsumed by `cmdproxy verify`.
The compatibility command should delegate to `verify` semantics rather than
maintain a separate execution path.

## Target Behavior

Rule-local examples still exist, but they are validated through `verify`.

## Scope

The compatibility command may continue to run rule-local examples, but user
documentation and hook workflows should direct users to `cmdproxy verify`.

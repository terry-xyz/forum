# Superpowers Collaboration Workflow

## Original Request

> i did the instructions and downloaded https://github.com/obra/superpowers/tree/main?search=1 . give me a complete guide on what it does and how to use it to collab with my collagues

## Practical Guide

I’m going to inspect the installed superpowers materials locally first, then I’ll give you a practical guide focused on collaboration workflows rather than a generic repo summary.

Using `using-superpowers` to follow the installed workflow, and I’m pulling the repo’s own docs now so the guide matches what you actually installed.

I have the top-level picture. Next I’m reading the collaboration-oriented skills so I can turn this into an actual team workflow you can follow with colleagues.

I’m checking your actual local install state now so I can tell you whether it’s just downloaded or also wired into Codex correctly.

`superpowers` is not an app framework or library. It is a workflow layer for coding agents: a set of `SKILL.md` instructions that Codex discovers at startup and applies automatically. In practice, it changes how the agent works so it stops jumping straight into code and instead follows a stricter flow: design first, plan second, implement with TDD, review often, finish branches cleanly.

### Install State

Your install looks mostly correct:

- The repo is present at `/home/student/.codex/superpowers`
- Codex can see the skills through the symlink at `/home/student/.agents/skills/superpowers`
- Your Codex config at `/home/student/.codex/config.toml` does not currently show `multi_agent = true`, so the parallel/subagent-heavy skills may be limited until you add:

```toml
[features]
multi_agent = true
```

### What It Does

- `using-superpowers`: forces Codex to check for relevant skills before acting.
- `brainstorming`: turns a rough idea into an approved design/spec before implementation.
- `writing-plans`: converts the approved spec into a detailed step-by-step implementation plan.
- `using-git-worktrees`: puts work into isolated branches/worktrees instead of contaminating your current checkout.
- `test-driven-development`: pushes RED -> GREEN -> REFACTOR instead of “write code and hope”.
- `subagent-driven-development` / `dispatching-parallel-agents`: splits independent tasks across agents when supported.
- `requesting-code-review` and `receiving-code-review`: adds structured review loops.
- `finishing-a-development-branch`: makes completion explicit: merge, PR, keep branch, or discard.

### How To Use It Yourself

Start a fresh Codex session after installing. Then prompt naturally; the skills should trigger from the task.

Examples:

- “Help me design a feature for X before coding it.”
- “Use brainstorming for this new feature.”
- “Write an implementation plan for this approved spec.”
- “Execute this plan.”
- “Review this change before I merge.”

### Normal Superpowers Flow

1. Idea: “I want to add feature X.”
2. Codex brainstorms and asks clarifying questions.
3. Codex writes a spec file.
4. You review and approve the spec.
5. Codex writes an implementation plan.
6. Codex implements in a worktree/branch, with tests and review checkpoints.
7. Codex offers finish options: merge, PR, keep branch, discard.

### How To Use It With Colleagues

Treat Superpowers as a team process, not just a personal plugin.

Recommended team workflow:

1. One person drives the brainstorming session with Codex.
2. Commit the generated spec and plan into the repo so teammates can read them.
3. Use one branch or worktree per feature/fix.
4. Have teammates review the spec before code starts, not just the PR after code exists.
5. Use the implementation plan as the shared contract for what gets built.
6. Keep agent-generated changes small and reviewable.
7. Use normal human code review on top of Superpowers review.

Good collaboration pattern:

- Product/teammate defines the goal.
- You ask Codex to brainstorm and write the spec.
- Team reviews spec in Git.
- You ask Codex to write the plan.
- Team checks plan for scope and sequencing.
- Codex implements task-by-task.
- Colleagues review PR against the spec and plan.

### Best Prompts For Collaboration

- “Use brainstorming and help me turn this idea into a spec my team can review.”
- “Write the spec to the repo so my colleagues can comment on it.”
- “Now write an implementation plan with small reviewable tasks.”
- “Implement only Task 1 from the plan in a separate branch.”
- “Request a code review against the plan before moving to the next task.”

### Important Caveats

- It is opinionated. It will slow down “just hack it quickly” workflows.
- It works best when specs and plans live in the repo.
- The strongest collaboration features depend on Codex multi-agent support.
- It does not replace human judgment; it mainly enforces discipline.

If you want, I can do the next useful step in this repo: verify your Codex setup fully and then give you a concrete “team playbook” file you can share with your colleagues.

<br>
<br>
<br>
<br>
<br>
<br>
<br>
<br>

## Three Features With Git Worktrees

> ok lets say i want to build 3 new features with git worktrees. what exactly should i do to be as quick and efficient as possible

For 3 new features, the fastest reliable setup is: one approved spec and one worktree per feature, with plans small enough that Codex can execute them without thrashing.

### Recommended Workflow

1. Start from a clean `main`.
2. For each feature, do a short spec first.
3. Create one worktree per feature branch.
4. Keep each feature isolated until it is merged.
5. Only run agents in the worktree for that feature.

### Practical Sequence

#### 1. Pick The Right Granularity

If the 3 features are independent, treat them as 3 separate tracks:

- `feature/a`
- `feature/b`
- `feature/c`

If they share the same internals or must land together, do one shared spec first, then split into implementation tasks. Do not split too early if the code overlaps heavily.

#### 2. Create A Lightweight Design For Each Feature

In Codex, for each feature, start with:

- “Use brainstorming to turn this feature into a short spec.”
- Review the spec.
- “Write the implementation plan.”

This matters because Superpowers is optimized for `spec -> plan -> implementation`. Skipping that usually costs time later.

#### 3. Use Worktrees So Nothing Blocks Anything Else

From your main repo:

```bash
git checkout main
git pull
mkdir -p .worktrees
git worktree add .worktrees/feature-a -b feature/a
git worktree add .worktrees/feature-b -b feature/b
git worktree add .worktrees/feature-c -b feature/c
```

Now you have:

- `/home/student/forum` as your main checkout
- `.worktrees/feature-a`
- `.worktrees/feature-b`
- `.worktrees/feature-c`

Each one is a separate branch with shared git objects, so it is much lighter than cloning 3 times.

#### 4. Run One Codex Session Per Worktree

Open a separate terminal in each worktree and keep each session focused on one feature only.

Example:

- terminal 1 -> `.worktrees/feature-a`
- terminal 2 -> `.worktrees/feature-b`
- terminal 3 -> `.worktrees/feature-c`

Then in each session:

- “Implement the approved plan for this feature.”
- Or: “Implement only Task 1 from the plan.”

That is the fastest pattern because each agent has narrow context and avoids cross-feature confusion.

#### 5. Prioritize Merge Order

Do this in order:

1. Merge the feature with the least dependency risk first.
2. Rebase or merge `main` into the others after that lands.
3. Keep PRs small.

If features touch the same files, parallel worktrees still help, but you should expect merge conflict cleanup. In that case, sequence the riskiest/shared-code feature first.

#### 6. Optimize For Speed

To move fastest:

- Keep each feature small enough to review in one PR.
- Ask Codex to implement one plan task at a time, not the whole universe.
- Use explicit prompts like:
  - “Implement only the first task.”
  - “Do not refactor unrelated code.”
  - “Request code review before moving on.”

#### 7. Finish Each Branch Cleanly

When a feature is done:

- run tests
- open PR or merge
- remove the worktree

Example:

```bash
git worktree remove .worktrees/feature-a
git branch -d feature/a
```

### Best Overall Playbook

- Feature 1: spec -> plan -> worktree -> implement -> review -> merge
- Feature 2: spec -> plan -> worktree -> implement -> review -> merge
- Feature 3: spec -> plan -> worktree -> implement -> review -> merge

### What Not To Do

- Don’t build all 3 features in one branch.
- Don’t let one Codex session hop across multiple worktrees/features.
- Don’t parallelize features that heavily edit the same code unless you accept conflict resolution overhead.
- Don’t skip the short spec/plan stage if you want efficiency; that is where most rework gets prevented.

<br>
<br>
<br>
<br>
<br>
<br>
<br>
<br>

## Step-By-Step Team Process

> ok i will explain how i am about to code with my collagues and correct me wherever i am wrong. I will first send what i want to build, then ask for clarifying questions to create a spec. then what do i do, go straight to implementation plan or do i split the spec into smaller ones to use git worktrees? give me the complete step by step process

After the clarifying questions and spec, do not automatically jump to the implementation plan.

The decision point is:

- If the spec describes one coherent feature that can be built on one branch, keep it as one spec and then go to the implementation plan.
- If the spec actually contains multiple independent features/subsystems, split it into separate smaller specs first. Then each smaller spec gets its own plan and usually its own git worktree.

The rule is simple: specs define scope; worktrees isolate implementation.
So you split by product/engineering independence, not just because you want multiple worktrees.

### Step-By-Step Process

1. Start with the product idea.
   Prompt Codex with something like:
   - “Use brainstorming. I want to build X.”
2. Let it ask clarifying questions.
   Answer until the goal, constraints, and success criteria are clear.
3. Review the proposed approaches.
   Superpowers should present 2-3 approaches with tradeoffs and a recommendation.
4. Approve the design/spec.
   Codex should write a spec document.
5. Decide whether the spec is too broad.
   Ask:
   - Can Feature A ship without Feature B?
   - Do different teammates own different parts?
   - Would these parts likely be separate PRs?
   - Do they touch mostly different files/systems?

   If mostly yes, split into separate specs.

   Example:
   - Bad single spec: “Build notifications, admin dashboard, and billing changes”
   - Better:
     - `notifications` spec
     - `admin-dashboard` spec
     - `billing-adjustments` spec
6. For each approved spec, create an implementation plan.
   Only now use `writing-plans`.
   Each plan should be narrow and executable.
7. Create one git worktree per plan/feature.
   Typical mapping:
   - 1 spec -> 1 plan -> 1 branch -> 1 worktree

   Example:

```bash
git checkout main
git pull
mkdir -p .worktrees
git worktree add .worktrees/notifications -b feature/notifications
git worktree add .worktrees/admin-dashboard -b feature/admin-dashboard
git worktree add .worktrees/billing -b feature/billing
```

8. Assign ownership with colleagues.
   Each colleague works in one worktree only.
   Do not have multiple people editing the same feature branch unless necessary.
9. In each worktree, run a separate Codex session.
   Use prompts like:
   - “Implement this approved plan.”
   - “Implement only Task 1.”
   - “Request code review before moving to the next task.”
10. Keep features isolated.
    Do not mix unrelated changes across worktrees.
    If one feature depends on another, merge the dependency first, then rebase/update the later branch.
11. Review at two levels.

- Human review of the spec before coding
- Human/agent review of the PR against the spec and plan

12. Merge one feature at a time.
    Prefer merging the least risky or most foundational feature first.
13. Sync remaining worktrees after merges.
    For each remaining feature branch:

```bash
git fetch origin
git merge origin/main
```

Or rebase if your team prefers that.

14. Finish and clean up.
    After merge:

```bash
git worktree remove .worktrees/notifications
git branch -d feature/notifications
```

### Best Structure For Collaborating With Colleagues

If you have 3 independent features, the ideal flow is:

1. Brainstorm Feature 1 -> approve spec -> write plan -> create worktree
2. Brainstorm Feature 2 -> approve spec -> write plan -> create worktree
3. Brainstorm Feature 3 -> approve spec -> write plan -> create worktree
4. Each person takes one feature/worktree
5. Merge in dependency order

### When Not To Split The Spec

Keep one spec if the work is tightly coupled, for example:

- one backend schema change and one UI change that are part of the same feature
- shared logic where splitting would create fake boundaries
- one feature that should land in one PR

In that case:

- one spec
- one plan
- one worktree
- maybe split the plan into tasks for people, but be careful about overlap

### Practical Rule Of Thumb

Use separate specs/worktrees when the answer to this is yes:

> “Could this be reviewed, tested, and merged without the other part?”

If yes, split it. If no, keep it together.

### Short Version

Your process should be:

1. Idea
2. Clarifying questions
3. Approve spec
4. Check if spec should be split
5. If yes: create smaller specs
6. For each final spec: write implementation plan
7. For each plan: create git worktree
8. One colleague per worktree
9. Implement, review, merge, clean up

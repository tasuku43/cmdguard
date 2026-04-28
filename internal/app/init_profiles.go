package app

type InitProfile struct {
	Name        string
	Description string
	Config      string
}

func InitProfiles() []InitProfile {
	return []InitProfile{
		{Name: "balanced", Description: "safe starter policy for common read-only operations", Config: balancedProfileConfig},
		{Name: "strict", Description: "minimal policy that asks for most operations and denies destructive examples", Config: strictProfileConfig},
		{Name: "git-safe", Description: "Git-focused policy for read-only work and safer history handling", Config: gitSafeProfileConfig},
		{Name: "aws-k8s", Description: "AWS and kubectl read-only starter policy", Config: awsK8sProfileConfig},
		{Name: "argocd", Description: "Argo CD read/status policy with destructive deletion blocked", Config: argocdProfileConfig},
	}
}

func InitProfileNames() []string {
	profiles := InitProfiles()
	names := make([]string, 0, len(profiles))
	for _, profile := range profiles {
		names = append(names, profile.Name)
	}
	return names
}

func LookupInitProfile(name string) (InitProfile, bool) {
	for _, profile := range InitProfiles() {
		if profile.Name == name {
			return profile, true
		}
	}
	return InitProfile{}, false
}

const balancedProfileConfig = `permission:
  deny:
    - name: git force push
      command:
        name: git
        semantic:
          verb: push
          force: true
      message: "git force push is blocked"
      test:
        deny:
          - "git push --force origin main"
        abstain:
          - "git push origin main"

    - name: git hard reset
      command:
        name: git
        semantic:
          verb: reset
          hard: true
      message: "git reset --hard is blocked"
      test:
        deny:
          - "git reset --hard HEAD"
        abstain:
          - "git reset HEAD"

  ask:
    - name: git write operations
      command:
        name: git
        semantic:
          verb_in:
            - commit
            - push
            - reset
      message: "git write operation requires confirmation"
      test:
        ask:
          - "git commit -m update"
          - "git push origin main"
        abstain:
          - "git status"

    - name: kubectl mutating operations
      command:
        name: kubectl
        semantic:
          verb_in:
            - apply
            - delete
            - scale
      message: "kubectl mutating operation requires confirmation"
      test:
        ask:
          - "kubectl apply -f app.yaml"
          - "kubectl delete pod app"
        abstain:
          - "kubectl get pods"

  allow:
    - name: git read-only
      command:
        name: git
        semantic:
          verb_in:
            - status
            - diff
            - log
            - show
      message: "allow git read-only commands"
      test:
        allow:
          - "git status"
          - "git diff --cached"
        abstain:
          - "git push origin main"

    - name: kubectl read-only
      command:
        name: kubectl
        semantic:
          verb_in:
            - get
            - describe
      message: "allow kubectl read-only commands"
      test:
        allow:
          - "kubectl get pods"
          - "kubectl describe pod app"
        abstain:
          - "kubectl delete pod app"

test:
  allow:
    - "git status"
    - "kubectl get pods"
  ask:
    - "git push origin main"
    - "kubectl apply -f app.yaml"
    - "unknown-tool --flag"
  deny:
    - "git push --force origin main"
    - "git reset --hard HEAD"
`

const strictProfileConfig = `permission:
  deny:
    - name: git force push
      command:
        name: git
        semantic:
          verb: push
          force: true
      message: "git force push is blocked"
      test:
        deny:
          - "git push --force origin main"
        abstain:
          - "git push origin main"

    - name: git hard reset
      command:
        name: git
        semantic:
          verb: reset
          hard: true
      message: "git reset --hard is blocked"
      test:
        deny:
          - "git reset --hard HEAD"
        abstain:
          - "git reset HEAD"

    - name: kubectl delete
      command:
        name: kubectl
        semantic:
          verb: delete
      message: "kubectl delete is blocked"
      test:
        deny:
          - "kubectl delete pod app"
        abstain:
          - "kubectl get pods"

  ask:
    - name: git read-only review
      command:
        name: git
        semantic:
          verb_in:
            - status
            - diff
            - log
            - show
      message: "git command requires confirmation in strict profile"
      test:
        ask:
          - "git status"
          - "git diff --cached"
        abstain:
          - "kubectl get pods"

    - name: kubectl read-only review
      command:
        name: kubectl
        semantic:
          verb_in:
            - get
            - describe
      message: "kubectl command requires confirmation in strict profile"
      test:
        ask:
          - "kubectl get pods"
          - "kubectl describe pod app"
        abstain:
          - "git status"

test:
  ask:
    - "git status"
    - "kubectl get pods"
    - "unknown-tool --flag"
  deny:
    - "git push --force origin main"
    - "git reset --hard HEAD"
    - "kubectl delete pod app"
`

const gitSafeProfileConfig = `permission:
  deny:
    - name: git force push
      command:
        name: git
        semantic:
          verb: push
          force: true
      message: "git force push is blocked"
      test:
        deny:
          - "git push --force origin main"
        abstain:
          - "git push origin main"

    - name: git hard reset
      command:
        name: git
        semantic:
          verb: reset
          hard: true
      message: "git reset --hard is blocked"
      test:
        deny:
          - "git reset --hard HEAD"
        abstain:
          - "git reset HEAD"

  ask:
    - name: git write operations
      command:
        name: git
        semantic:
          verb_in:
            - commit
            - push
            - reset
            - rebase
            - merge
      message: "git write operation requires confirmation"
      test:
        ask:
          - "git commit -m update"
          - "git rebase main"
        abstain:
          - "git status"

  allow:
    - name: git read-only
      command:
        name: git
        semantic:
          verb_in:
            - status
            - diff
            - log
            - show
            - branch
      message: "allow git read-only commands"
      test:
        allow:
          - "git status"
          - "git diff --cached"
          - "git log --oneline"
        abstain:
          - "git push origin main"

test:
  allow:
    - "git status"
    - "git diff --cached"
    - "git log --oneline"
  ask:
    - "git commit -m update"
    - "git push origin main"
    - "git rebase main"
  deny:
    - "git push --force origin main"
    - "git reset --hard HEAD"
`

const awsK8sProfileConfig = `permission:
  ask:
    - name: AWS non-read operations
      command:
        name: aws
        semantic:
          operation_in:
            - delete
            - put
            - update
            - create
            - terminate-instances
      message: "AWS mutating operation requires confirmation"
      test:
        ask:
          - "aws ec2 terminate-instances --instance-ids i-1234567890abcdef0"
        abstain:
          - "aws sts get-caller-identity"

    - name: kubectl mutating operations
      command:
        name: kubectl
        semantic:
          verb_in:
            - apply
            - delete
            - scale
            - rollout
      message: "kubectl mutating operation requires confirmation"
      test:
        ask:
          - "kubectl apply -f app.yaml"
          - "kubectl delete pod app"
        abstain:
          - "kubectl get pods"

  allow:
    - name: AWS identity
      command:
        name: aws
        semantic:
          service: sts
          operation: get-caller-identity
      message: "allow AWS identity check"
      test:
        allow:
          - "aws sts get-caller-identity"
        abstain:
          - "aws s3 ls"

    - name: AWS read-only examples
      command:
        name: aws
        semantic:
          operation_in:
            - describe-instances
            - list-buckets
            - get-caller-identity
      message: "allow selected AWS read-only commands"
      test:
        allow:
          - "aws ec2 describe-instances"
          - "aws s3api list-buckets"
        abstain:
          - "aws ec2 terminate-instances --instance-ids i-1234567890abcdef0"

    - name: kubectl read-only
      command:
        name: kubectl
        semantic:
          verb_in:
            - get
            - describe
            - logs
      message: "allow kubectl read-only commands"
      test:
        allow:
          - "kubectl get pods"
          - "kubectl describe pod app"
        abstain:
          - "kubectl delete pod app"

test:
  allow:
    - "aws sts get-caller-identity"
    - "aws ec2 describe-instances"
    - "kubectl get pods"
  ask:
    - "aws ec2 terminate-instances --instance-ids i-1234567890abcdef0"
    - "kubectl apply -f app.yaml"
    - "kubectl delete pod app"
`

const argocdProfileConfig = `permission:
  deny:
    - name: Argo CD app delete
      command:
        name: argocd
        semantic:
          verb: app delete
      message: "Argo CD app deletion is blocked"
      test:
        deny:
          - "argocd app delete my-app"
        abstain:
          - "argocd app get my-app"

  ask:
    - name: Argo CD app sync
      command:
        name: argocd
        semantic:
          verb: app sync
      message: "Argo CD app sync requires confirmation"
      test:
        ask:
          - "argocd app sync my-app"
        abstain:
          - "argocd app get my-app"

  allow:
    - name: Argo CD app read
      command:
        name: argocd
        semantic:
          verb_in:
            - app get
            - app list
      message: "allow Argo CD app read commands"
      test:
        allow:
          - "argocd app get my-app"
          - "argocd app list"
        abstain:
          - "argocd app sync my-app"

test:
  allow:
    - "argocd app get my-app"
    - "argocd app list"
  ask:
    - "argocd app sync my-app"
  deny:
    - "argocd app delete my-app"
`

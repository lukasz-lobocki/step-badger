package main

var (
	semVer      string
	commitHash  string
	isGitDirty  string
	isSnapshot  string
	goOs        string
	goArch      string
	gitUrl      string
	builtBranch string
	builtDate   string
)

var semReleaseVersion string = semVer +
	func(pre string, txt string) string {
		if len(txt) > 0 {
			return pre + txt
		} else {
			return ""
		}
	}("+", goArch) +
	func(pre string, txt string) string {
		if len(txt) > 0 {
			return pre + txt
		} else {
			return ""
		}
	}(".", builtBranch) +
	func(pre string, txt string) string {
		if len(txt) > 0 {
			return pre + txt
		} else {
			return ""
		}
	}(".", commitHash)

func main() {
	println("release version:", semReleaseVersion)
	println()
	println("semVer:", semVer)
	println("commitHash:", commitHash)
	println("isGitDirty:", isGitDirty)
	println("isSnapshot:", isSnapshot)
	println("goOs:", goOs)
	println("goArch:", goArch)
	println("gitUrl:", gitUrl)
	println("builtBranch:", builtBranch)
	println("builtDate:", builtDate)
}

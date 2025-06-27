package test

import (
	"testing"

	"gman/internal/git"
)

// Benchmark comparing hybrid manager (go-git enabled) vs native git only
func BenchmarkHybridVsNative(b *testing.B) {
	// Use current repository for testing
	repos := map[string]string{
		"cli-tool": "/Users/henrykuo/tui/programming/cli-tool",
	}

	b.Run("HybridManager-GoGitEnabled", func(b *testing.B) {
		hybridMgr := git.NewHybridManager()
		hybridMgr.SetGoGitEnabled(true)
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for alias, path := range repos {
				_ = hybridMgr.GetRepoStatusNoFetch(alias, path)
			}
		}
	})

	b.Run("HybridManager-GoGitDisabled", func(b *testing.B) {
		hybridMgr := git.NewHybridManager()
		hybridMgr.SetGoGitEnabled(false) // Use native git only
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for alias, path := range repos {
				_ = hybridMgr.GetRepoStatusNoFetch(alias, path)
			}
		}
	})

	b.Run("NativeManager", func(b *testing.B) {
		nativeMgr := git.NewManager()
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for alias, path := range repos {
				_ = nativeMgr.GetRepoStatusNoFetch(alias, path)
			}
		}
	})
}

// Benchmark single repository status checking performance
func BenchmarkSingleRepoStatus(b *testing.B) {
	// Use current repository for testing
	testRepo := "cli-tool"
	testPath := "/Users/henrykuo/tui/programming/cli-tool"

	b.Run("GoGit-StatusReader", func(b *testing.B) {
		nativeMgr := git.NewManager()
		goGitStatus := git.NewGoGitStatusReader(nativeMgr)
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = goGitStatus.GetRepoStatusNoFetch(testRepo, testPath)
		}
	})

	b.Run("Native-StatusReader", func(b *testing.B) {
		nativeMgr := git.NewManager()
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = nativeMgr.GetRepoStatusNoFetch(testRepo, testPath)
		}
	})
}

// Benchmark workspace status checking (most expensive operation)
func BenchmarkWorkspaceStatus(b *testing.B) {
	// Use current repository for testing
	testPath := "/Users/henrykuo/tui/programming/cli-tool"

	b.Run("GoGit-HasUncommittedChanges", func(b *testing.B) {
		nativeMgr := git.NewManager()
		goGitStatus := git.NewGoGitStatusReader(nativeMgr)
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = goGitStatus.HasUncommittedChanges(testPath)
		}
	})

	b.Run("Native-HasUncommittedChanges", func(b *testing.B) {
		nativeMgr := git.NewManager()
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = nativeMgr.HasUncommittedChanges(testPath)
		}
	})
}
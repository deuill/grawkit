# Grawkit - The Awksome Git Graph Generator

Grawkit is a tool that helps build SVG graphs from git command-line descriptions, and is built in Awk.

## Testing & Documentation

A `Makefile` is provided for running tests and producing documentation for Grawkit. Run `make help` in the project root for more information.

A full test-suite is provided (depending only on `make` and `awk`), which should serve as a good example of the existing feature-set.

## Status & Examples

Grawkit has basic support for common `git` commands such as `git branch`, `git tag` and `git merge`, allowing for fairly complex graphs. The integrated test-suite serves as an example, presented here:

<table>
	<tr>
		<th width="40%">Command-Line</th>
		<th>Generated Graph</th>
	</tr>
	<tr>
		<th><pre><code>git commit -m "Adding a new commit"
git commit</code></pre></th>
		<th><img src="https://rawgit.com/deuill/grawkit/55cacf1c890c47f99a3e9c9dbd86a8fffe464568/tests/02-master.svg" alt=""></th>
	</tr>
	<tr>
		<th><pre><code>git commit -m "Commit on master"
git commit -m "More stuff"

git branch test-stuff
git checkout test-stuff

git commit -m 'Testing stuff'
git commit

git checkout master
git commit</code></pre></th>
		<th><img src="https://rawgit.com/deuill/grawkit/55cacf1c890c47f99a3e9c9dbd86a8fffe464568/tests/03-branch.svg" alt=""></th>
	</tr>
	<tr>
		<th><pre><code>git branch test-merging

git commit -m "Commit on master"
git commit -m "Still on master"

git checkout test-merging
git commit -m 'A sample commit'

git checkout master
git commit -m "Another master commit"

git merge test-merging</code></pre></th>
		<th><img src="https://rawgit.com/deuill/grawkit/55cacf1c890c47f99a3e9c9dbd86a8fffe464568/tests/04-merge.svg" alt=""></th>
	</tr>
	<tr>
		<th><pre><code>git commit -m "Commit on master"

git branch test-first
git branch test-second

git commit -m "Still on master"
git tag v.1.0.0

git checkout test-first
git commit

git branch test-third

git checkout test-second
git commit
git merge test-first
git tag v.2.0.0-rc1

git checkout master
git merge test-second

git checkout test-third
git commit</code></pre></th>
		<th><img src="https://rawgit.com/deuill/grawkit/55cacf1c890c47f99a3e9c9dbd86a8fffe464568/tests/05-multi-branch.svg" alt=""></th>
	</tr>
	<tr>
		<th><pre><code>git branch develop
git checkout develop
git commit

git branch feature/XYZ-12_fix-foo
git checkout feature/XYZ-12_fix-foo
git commit
git commit

git checkout develop
git branch feature/ZZ-704_take-it-to-the-limit
git checkout feature/ZZ-704_take-it-to-the-limit
git commit

git checkout develop
git merge feature/XYZ-12_fix-foo

git branch feature/ABC-66_make-bar
git checkout feature/ABC-66_make-bar
git commit
git commit

git checkout develop
git merge feature/ZZ-704_take-it-to-the-limit</code></pre></th>
		<th><img src="https://rawgit.com/deuill/grawkit/55cacf1c890c47f99a3e9c9dbd86a8fffe464568/tests/06-feature-branch.svg" alt=""></th>
	</tr>
</table>

## License

All code in this repository is covered by the terms of the MIT License, the full text of which can be found in the LICENSE file.

[license-url]: https://github.com/deuill/grawkit/blob/master/LICENSE
[license-svg]: https://img.shields.io/badge/license-MIT-blue.svg

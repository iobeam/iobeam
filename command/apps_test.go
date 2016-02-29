package command

import "testing"

// TestLaunchAppArgsIsValid tests validity cases for launchAppArgs
func TestLaunchAppArgsIsValid(t *testing.T) {
	cases := []dataTestCase{
		{
			desc: "a valid launchAppArgs object",
			in: &launchAppArgs{
				uploadFileArgs: uploadFileArgs{
					projectId: 1,
					path:      "test/path.out",
				},
				name: "test-app",
			},
			want: true,
		},
		{
			desc: testDescInvalidProjectId,
			in: &launchAppArgs{
				uploadFileArgs: uploadFileArgs{
					projectId: 0,
					path:      "test/path.out",
				},
				name: "test-app",
			},
			want: false,
		},
		{
			desc: "invalid path given (none)",
			in: &launchAppArgs{
				uploadFileArgs: uploadFileArgs{
					projectId: 1,
					path:      "",
				},
				name: "test-app",
			},
			want: false,
		},
		{
			desc: "invalid name given (none)",
			in: &launchAppArgs{
				uploadFileArgs: uploadFileArgs{
					projectId: 1,
					path:      "test/path.out",
				},
				name: "",
			},
			want: false,
		},
	}
	runDataTestCase(t, cases)
}

// TestBaseAppArgsIsValid tests validity cases for baseAppArgs
func TestBaseAppArgsIsValid(t *testing.T) {
	cases := []dataTestCase{
		{
			desc: "a valid baseAppArgs object w/ both id & name",
			in: &baseAppArgs{
				projectId: 1,
				name:      "test-app",
				id:        1,
			},
			want: true,
		},
		{
			desc: "a valid baseAppArgs object w/ name",
			in: &baseAppArgs{
				projectId: 1,
				name:      "test-app",
			},
			want: true,
		},
		{
			desc: "a valid baseAppArgs object w/ id",
			in: &baseAppArgs{
				projectId: 1,
				id:        1,
			},
			want: true,
		},
		{
			desc: testDescInvalidProjectId,
			in: &baseAppArgs{
				projectId: 0,
				id:        1,
			},
			want: false,
		},
		{
			desc: "invalid, no id or name",
			in: &baseAppArgs{
				projectId: 1,
			},
			want: false,
		},
	}
	runDataTestCase(t, cases)
}

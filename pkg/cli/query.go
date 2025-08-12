package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/ast9501/tq/pkg/tq"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

// Query does the following:
// 1. Converts incoming Terraform bytes to JSON
// 2. Runs it through jq (via subprocess) to apply the query
// 3. Converts it back to Terraform (Note: if this fails, just return the JSON)
func Query(argMap ArgMap, query string, in []byte) ([]byte, error) {
	// Convert original input to JSON
	tfFile, err := tq.ParseTerraform(in)
	if err != nil {
		return nil, err
	}
	jb, err := json.Marshal(tfFile)
	if err != nil {
		return nil, err
	}

	var cmd *exec.Cmd
	args := []string{"jq"}
	if len(argMap) > 0 {
		for key, value := range argMap {
			// force wrap value with "
			args = append(args, "--arg", key, "\""+value+"\"")
		}
		args = append(args, "-r", query)
		cmd = exec.Command(args[0], args[1:]...)
	} else {
		// Run jq directly, passing the parsed JSON as STDIN
		cmd = exec.Command("jq", "-r", query)
	}
	var jqbuff bytes.Buffer
	cmd.Stdout = &jqbuff
	cmd.Stderr = os.Stderr
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	if _, err := io.WriteString(stdin, string(jb)); err != nil {
		return nil, err
	}
	stdin.Close()
	if err := cmd.Wait(); err != nil {
		return nil, err
	}

	// Now attempt to convert it back to tf
	jqb := jqbuff.Bytes()
	file, err := tq.ParseJSON(jqb)
	if err != nil {
		// TODO: add option to fail here
		// If there was some issue converting back to terraform,
		// just print out the JSON (assume what the user wants here)
		return jqb, nil
	}
	tfb := hclwrite.Format(file.Bytes())
	if tfb == nil {
		// TODO: add option to fail here
		// Like above, maybe we somehow passed conversion, but
		// the result is empty. Return the jq raw output.
		return jqb, nil
	}

	return tfb, nil
}

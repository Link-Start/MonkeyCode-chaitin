import assert from "node:assert/strict"
import { readFile } from "node:fs/promises"
import test from "node:test"

test("models page uses backend models and authorization subjects", async () => {
  const source = await readFile(
    new URL("../src/pages/models-page.tsx", import.meta.url),
    "utf8"
  )

  assert.doesNotMatch(source, /INITIAL_MODELS/)
  assert.match(source, /\/api\/admin\/v1\/models/)
  assert.match(source, /\/authorization-subjects/)
  assert.match(source, /max_output_tokens/)
  assert.match(source, /api_key_configured/)
  assert.match(source, /openai_responses/)
  assert.doesNotMatch(source, /member-01|engineering|operations/)
})

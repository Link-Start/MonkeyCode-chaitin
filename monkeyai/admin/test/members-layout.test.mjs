import assert from "node:assert/strict"
import { readFile } from "node:fs/promises"
import test from "node:test"

test("members page keeps the original split cards and compact member list", async () => {
  const source = await readFile(
    new URL("../src/pages/members-and-groups-page.tsx", import.meta.url),
    "utf8"
  )

  assert.match(source, /md:grid-cols-\[minmax\(14rem,1fr\)_minmax\(0,2fr\)\]/)
  assert.match(source, /pages\.membersAndGroups\.groupsTitle/)
  assert.match(source, /<ItemGroup/)
  assert.match(source, /<ItemFooter/)
  assert.match(source, /<DropdownMenu/)
  assert.doesNotMatch(source, /<Table/)
})

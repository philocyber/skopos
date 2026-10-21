import assert from "node:assert/strict";
import test from "node:test";
import { auth } from "./auth.ts";

function browser() {
  const storage = new Map();
  const cookies = new Map();
  const writes = [];
  globalThis.window = {};
  globalThis.localStorage = {
    getItem: (key) => storage.get(key) ?? null,
    setItem: (key, value) => storage.set(key, value),
    removeItem: (key) => storage.delete(key),
  };
  globalThis.document = {
    get cookie() { return [...cookies].map(([key, value]) => `${key}=${value}`).join("; "); },
    set cookie(value) {
      writes.push(value);
      const [name, entry] = value.split(";")[0].split("=");
      if (value.includes("max-age=0")) cookies.delete(name);
      else cookies.set(name, entry);
    },
  };
  return { storage, cookies, writes };
}

test("login and password refresh never put credentials in a cookie", () => {
  const state = browser();
  for (const token of ["initial.secret.jwt", "replacement.secret.jwt"]) {
    auth.setToken(token);
    assert.equal(auth.getToken(), token);
    assert.equal(state.storage.get("skopos_token"), token);
    assert.equal(state.cookies.get("skopos_session"), "1");
    assert.equal(state.cookies.has("skopos_token"), false);
    assert.ok(state.writes.every((write) => !write.includes(token)));
  }
  auth.clearToken();
  assert.equal(state.storage.has("skopos_token"), false);
  assert.equal(state.cookies.has("skopos_session"), false);
});

test("existing localStorage sessions survive legacy cookie removal", () => {
  const state = browser();
  state.storage.set("skopos_token", "existing.secret.jwt");
  state.cookies.set("skopos_token", "existing.secret.jwt");
  assert.equal(auth.getToken(), "existing.secret.jwt");
  assert.equal(state.cookies.has("skopos_token"), false);
  auth.setToken(auth.getToken());
  assert.equal(state.cookies.get("skopos_session"), "1");
});

test("forged page-routing hint does not create a client credential", () => {
  const state = browser();
  state.cookies.set("skopos_session", "1");
  assert.equal(auth.getToken(), null);
  auth.clearToken();
  assert.equal(state.cookies.size, 0);
});

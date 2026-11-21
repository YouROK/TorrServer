/* eslint-disable */
// Copies the files of the JASSUB library (libass in WebAssembly, used to show ASS subtitles
// with full styling) from node_modules to public/jassub, so that they end up in the web build
// and in the embedded pages. Installed with: yarn add jassub@~1.8
// Without the library the player uses its simple ASS renderer.
const fs = require('fs')
const path = require('path')

const source = path.join(__dirname, '..', 'node_modules', 'jassub', 'dist')
const target = path.join(__dirname, '..', 'public', 'jassub')

if (!fs.existsSync(source)) {
  console.warn('[copy-jassub] node_modules/jassub is not installed, skipped (run "yarn add jassub@~1.8")')
  process.exit(0)
}

fs.rmSync(target, { recursive: true, force: true })
fs.mkdirSync(target, { recursive: true })

const wanted = /\.(js|wasm|woff2?|ttf)$/i
const unwanted = /\.(es|esm|mjs|map)\.|\.d\.ts$|\.map$|\.es\.js$/i
let count = 0
fs.readdirSync(source).forEach(name => {
  const file = path.join(source, name)
  if (!fs.statSync(file).isFile() || !wanted.test(name) || unwanted.test(name)) return
  fs.copyFileSync(file, path.join(target, name))
  count++
})
console.log(`[copy-jassub] ${count} file(s) copied to public/jassub`)

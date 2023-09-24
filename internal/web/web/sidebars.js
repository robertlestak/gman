/**
 * Creating a sidebar enables you to:
 - create an ordered group of docs
 - render a sidebar for each doc of that group
 - provide next/previous navigation

 The sidebars can be generated from the filesystem, or explicitly defined here.

 Create as many sidebars as you want.
 */
const fs = require('fs');
const path = require('path');

function walk(dir, callback) {
  fs.readdir(dir, (err, files) => {
      if (err) return callback(err);

      files.forEach((file) => {
          const fullPath = path.join(dir, file);
          fs.stat(fullPath, (err, stats) => {
              if (err) return callback(err);

              if (stats.isDirectory()) {
                  walk(fullPath, callback);
              } else if (stats.isFile()) {
                  callback(null, fullPath);
              }
          });
      });
  });
}

function walkSync(dir) {
  let fileList = [];

  const files = fs.readdirSync(dir);
  for (const file of files) {
      const fullPath = path.join(dir, file);
      const stats = fs.statSync(fullPath);

      if (stats.isDirectory()) {
          fileList = fileList.concat(walkSync(fullPath));
      } else if (stats.isFile()) {
          fileList.push(fullPath);
      }
  }

  return fileList;
}

function docsSidebar() {
  let items = walkSync(process.env.DOCS_DIR);
  // folders are in the structure of:
  // docs/{namespace}/{app}/*.md
  // we want to create a category for each namespace
  // and a sidebar item for each app
  const namespaces = new Set();
  const apps = {};

  items.forEach(filename => {
    let fn = filename.replace(process.env.DOCS_DIR, '');
    // split on slash
    let parts = fn.split('/');
    let namespace = parts[1];
    let app = parts[2];
    let doc = parts[3];
    // if doc doesn't have .md extension, skip it
    if (!doc || !doc.endsWith('.md')) {
      return;
    }
    let docName = doc.replace('.md', '');
    namespaces.add(namespace);
    if (!apps[namespace]) {
      apps[namespace] = [];
    }
    if (apps[namespace][app] === undefined) {
      apps[namespace][app] = {};
    }
    // remove / prefix
    fn = fn.replace('/', '');
    // remove .md suffix
    fn = fn.replace('.md', '');
    apps[namespace][app][docName] = fn;
  });
  let sidebar = [];
  // if there was a README.md file in the root of the docs dir,
  // set that as the index
  let indexFile = path.join(process.env.DOCS_DIR, 'README.md');
  if (fs.existsSync(indexFile)) {
    // add index file to sidebar
    sidebar.push(indexFile.replace(process.env.DOCS_DIR, '').replace("/", "").replace('.md', ''));
  }
  namespaces.forEach(namespace => {
    let category = {
      type: 'category',
      label: namespace,
      items: [],
    };
    for (const [app, docs] of Object.entries(apps[namespace])) {
      let item = {
        type: 'category',
        label: app,
        items: [],
      };
      for (const [docName, fn] of Object.entries(docs)) {
        item.items.push({
          type: 'doc',
          id: fn,
          label: docName,
        });
      }
      category.items.push(item);
    }
    sidebar.push(category);
  });
  return sidebar;
}

// @ts-check

/** @type {import('@docusaurus/plugin-content-docs').SidebarsConfig} */
const sidebars = {
  // By default, Docusaurus generates a sidebar from the docs folder structure
  docsSidebar: docsSidebar(),

  // But you can create a sidebar manually
  /*
  tutorialSidebar: [
    'intro',
    'hello',
    {
      type: 'category',
      label: 'Tutorial',
      items: ['tutorial-basics/create-a-document'],
    },
  ],
   */
};

module.exports = sidebars;

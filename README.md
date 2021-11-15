# iGenDec

iGenDec documentation is available from https://github.com/blgolden/igendec/wiki including the installation instructions.

### EPDs structure

Tree structure that goes to as much depth as you would like. Recommend only going two levels deep, otherwise the path will be too long for the interface.

See the `epds/AHA` directory for an example structure. The icon is in the `epds/AHA/icon.png` file, so applies to each of the databases (`2018Bulls`, `2019Bulls`, `2020Bulls`). If you add an icon to any of the databases (eg: `epds/AHA/2019Bulls/icon.png`), it will instead of the parent icon. If no icon is found it won't be displayed on the website.

### User Database Access Control

Example of a typical `hjson` user profile you will find in the `users` directory.

```
{
  Firstname: John
  // ... other details
  Access:
  [
    {
      Path: *
      Deny: false // optional, only required if setting true
    }
  ]
}
```

Where the Access field is used to control access to the databases. It is an array of objects where `Path` is the database path relative to the epds root, and `Deny` is optional, and if true that path will be denied access.

Path can include wildcards, eg: `AHA/*` to allow access to all AHA database paths. When a user first signs in, they are given access to all databases (so they have the permission you see in the example), it is to the admin to then refine the access.

When the algorithm checks if a user can view a database, it will search through the Access list checking for the best match. The best match is defined by the longest path with the least number of wildcards. For example, if the database in question is `AHA/2019Bulls` and the user has the following access:

```
  [
    {
      Path: *
    },
    {
      Path: AHA/*
      Deny: true
    }
    {
      Path: AHA/2019Bulls
    }
  ]
```

The user would be permitted access. The first policy allows access to all databases, the second policy denies access to any `AHA` databases (overriding the first policy as the path is more specific), and the third policy then allows `AHA/2019Bulls` only. The algoritm would find the last policy because its the best match to the database in question, see that `Deny == false` and allow access.

The algorithm does allow for more complex systems. Say if you wanted to make a small database available for every section in the structure, you could name the database `Sample` (so you would have `AHA/Sample`, `ASA/Sample`, `UNLAAA/Sample` etc), then you could use a policy like:

```
  [
    {
      Path: *
      Deny: true
    },
    {
      Path: */Sample
      Deny: false
    }
  ]
```
which would deny access to all datasets except a Sample database nested one level down.

### Dev Notes:

#### Performance:

- Each time a user creates a new run the default parameter files are re-read, parsed, and then injected into the html templates. If the server comes under a large stress this could affect the performance. It is helpful at early stages as it allows updating of the default files without a server reset, and it's simple but could update to parse once on start and keep the defaults sitting in memory.

#### Front end

- Uses html templates. The templates can be found in the views directory. The files html files in the top level are for each seperate end point a user can visit. All of the html files in the folders are the partials/components that will be updated with AJAX.
- The one exception to that is the 'layout' folder. This contains primary.html which has the general page layout (navbar, footer etc).
- Most of the business logic is found in the create folder. Each file in the create folder should directly correspond with a tab on the create page, with the exception of herdcomp, this is used for both Herd Comp and Bull Comp tabs as the logic is exactly the same with the exception of a few semantics. You will find some oddities in the herdcomp file. After each unique identifier I need to inject a suffix seen as {{.ID}} to keep the identifiers unique when the template is rendered twice on the front end. The pattern for the javascript is common logic should be found in the script of the create-build.html file, and tab specific javascript is contained in the html component. There is some cases where logic needs to be shared accross tabs, where previous choices redirect the flow of the interface

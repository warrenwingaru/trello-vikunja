# Trello Archives to Vikunja
Collection of utilities to export and import trello archived cards to Vikunja

## Usage
### Pre-requisites
Download the `exporter` and `migrator`

Export your vikunja data first. it must have a `data.json`. Which will be used to map Vikunja projects to Trello Boards

Your directory should look like this
```
data.json
exporter
migrator
```


### Exporter
modify the contents of [.env.example](.env.example) 
```
TRELLO_API_KEY=
TRELLO_API_TOKEN=
```

Run the application
```bash
./exporter # unix

./exporter.exe # windows

```

This will create a json file called `trello.json` where you can review the list of boards to export to Vikunja

hence the directory tree will now be as follows
```
data.json
trello.json
exporter
migrator
```

### Migrator
same as the exporter modify the contents of [.env.example](.env.example)
```
TRELLO_API_KEY=
TRELLO_API_TOKEN=
VIKUNJA_INSTANCE= # https://vikunja.tld/api/v1
```

Run the application
```bash
./migrator # inux
./migrator.exe # windows
```

This will read both `data.json` and `trello.json` and will ask you to choose which boards to perform the migration.

After choosing the numbers for the boards you want to migrate press enter.
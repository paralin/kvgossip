Create a Grant
-------------

`./kvgossip control --private root.pem newgrant --target entity_pub.pem --grantpattern "/fusebot.io/**"`

Set/Get a Key
---------

- `./kvgossip control setkey --key "/fusebot.io/r/np1" --value entity.pem`
- `./kvgossip control getkey --key "/fusebot.io/r/np1" --value entity.pem`

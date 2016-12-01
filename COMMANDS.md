Create a Grant
-------------

`./kvgossip control --private root.pem newgrant --target entity_pub.pem --grantpattern "/fusebot.io/**"`

Set a Key
---------

`./kvgossip --dbpath "kvgossip2.db" control setkey --key "/fusebot.io/r/np1" --value entity.pem`

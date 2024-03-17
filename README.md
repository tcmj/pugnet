# pugnet
Hyperlink Collection Tool



## Technologies
- [DuckDb ](https://duckdb.org) as a fast in-process analytical database
- [Zero Allocation JSON Logger ](https://github.com/rs/zerolog) 



Regex to match URLs:
```
^(?:([A-Za-z]+):)?(\/{0,3})([0-9.\-A-Za-z]+)(?::(\d+))?(?:\/([^?#]*))?(?:\?([^#]*))?(?:#(.*))?$
```

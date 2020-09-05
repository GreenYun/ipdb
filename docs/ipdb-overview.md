IPDB file
=========

This specification describes the structure of IPDB file provided by [ipip.net](https://www.ipip.net). This is not an official document but something referential, and there is no guarantee of the accuracy.
To deal with IPDB files, it is recommended to do with the [code from official support](https://www.ipip.net/support/code.html).

## Overview

The file can be simply divided in to two main parts:

- [File Header](#file-header)
   - Size indicator
   - [Metadata](#metadata)
- [Raw data](#raw-data)
   - [Nodes](#nodes)
   - [Leafs](#leafs)

## File Header

At the beginning of an IPDB file is a File Header in the following format. 

| Offset | Size           | Field          | Type           | Description           |
|--------|----------------|----------------|----------------|-----------------------|
| 0      | 4              | SizeOfMetadata | 32-bit Integer | The size of Metadata. |
| 4      | SizeOfMetadata | Metadata       | Byte Array     | The Metadata.         |

### Metadata

Metadata comes with a byte stream, or simply a string, that contains just a JSON object. Here is an example of Metadata:

```JSON
{
    "build": 1535696240,
    "ip_version": 1,
    "languages": {
        "CN": 0,
        "EN": 3
    },
    "node_count": 385083,
    "total_size": 3117287,
    "fields": [
        "country_name",
        "region_name",
        "city_name"
    ]
}
```

| Field      | Type                 | Description                                                                                                                           |
|------------|----------------------|---------------------------------------------------------------------------------------------------------------------------------------|
| build      | Integer              | Unix time when the file was generated.                                                                                                |
| ip_version | Integer              | The number identifies the IP version supported by the current database. `1` means IPv4, `2` means IPv6; others are not defined yet.   |
| languages  | String-Integer Pairs | The language code and the magic number pair. See [leafs](#leafs) for details.                                                         |
| node_count | Integer              | See [nodes](#nodes).                                                                                                                  |
| total_size | Integer              | The size of Raw Data. Useful to do completeness verification of the file.                                                             |
| fields     | String Array         | While the data content was split into fields, the field names are shown right here, with the exactly same order as the data contents. |

## Raw Data

Immediately after the File Header is the Raw Data. One may consider it as one large array plus one byte stream in RAM, of respectively `node` and `leaf`, which can be described as

```C
// struct Node in C
struct node {
    uint32_t zero, one;
}

// struct Leaf in C (without 4-byte alignment)
struct leaf {
    uint16_t size;
    char content[];  // Unlike C, content here is not a pointer but the array itself
}
```

### `node`s

The amount of `node` is defined as `metadata.node_count` in Metadata, which indicates the length of the `node` array, of index between `0..metadata.node_count-1`. Apparently, every element sizes exactly 8 bytes in the array. Either the value of `zero` or `one` of a `node` is the index of the `node` array or the "*offset*" to get a `leaf`, like a pointer pointing to its child element. 
We will mention the "*offset*" that points to a `leaf` below.

Therefore, the whole data constructs a tree-like structure:

```
              root
        ---------------
       |               |
       0               1         First bit of an IP
    -------         -------
   |       |       |       |
   0       1       0       1     Second bit of an IP
  ---     ---     ---     ---
 |   |   |   |   |   |   |   |
 0   1   0   1   0   1   0   1   Third bit of an IP
 -   -   -   -   -   -   -   -
| | | | | | | | | | | | | | | |
0 1 0 1 0 1 0 1 0 1 0 1 0 1 0 1  Fourth bit of an IP
...............................  ...
```

No matter the database is of IPv4 or IPv6, we treat any IP addresses as 128-bit binary data, by converting the IPv4 into an IPv4-Mapped IPv6 uni-cast address. E.g. to examine some information about `8.8.8.8`, we translate it into IPv6 address `::ffff:0808:0808`. While some `node`s may not exist in an IPv4 database, we should not try any other than an IPv4-Mapped IPv6 uni-cast address for examining.

The parser program will start by set `index = 0`, then get a `node`, set `index = node.zero` if the first bit is `0`, otherwise `index = node.one`; then get a `node` again, decide `index` by looking at the next bit. When getting a `node`, the program multiplies `index` by 8, and the product shall be the actual offset from the **very beginning of the Raw Data**. Repeat this until you found `index >= metadata.node_count`, which means it leads you to a `leaf` that has the data you want. There is always no need to go through all 128 level to get a `leaf`, and if you get the `leaf` on level 120, for instance, the data satisfies all IP addresses in the segment `1:2:3:4:5:6:7:800/120` or `8.8.8.0/24`, etc. 

### `leaf`s

A byte stream (or a byte array) immediately follows the `node` array, containing every `leaf`. While one gets a number as "*offset*" (as mentioned before) to find a `leaf`, that is not the real offset. First of all, the number must be subtracted by `metadata.node_count`, resulting the offset from the start of all `leaf`s (the byte stream after the `node` array, **NOT THE WHOLE Raw Data**). Obviously, if we count from the very beginning of the Raw Data, the result must be added by `metadata.node_count * 8`.

The size of each `leaf` is determined by its first two bytes, which indicates the size of `leaf.content`, i.e. the total size of a `leaf` is `leaf.size + 2`. The raw content is a byte array or a string, containing data of each fields, listed in `metadata.fields`, in every language defined in `metadata.languages`, separated by a `9` of 8-bit integer or an ASCII `TAB` character (`'\t'`). We usually break the content into a string array, with separators abandoned. As mentioned before, the magic numbers of every language refer to the start index in this string array of the particular language. E.g. the content:

```
"美国\t加利福尼亚州\t山景城\tUS\tCA\tMountain View"
```

After dividing:

```
[0]"美国" [1]"加利福尼亚州" [2]"山景城" [3]"US" [4]"CA" [5]"Mountain View"
```

If we want to get only the English contents, and get the magic number `3` corresponding to language code `"EN"`, we just start extracting the contents from `[3]`, which is corresponding to the first element of `metadata.fields` as `country_name`; and go on for each in `metadata.fields` in order. Finally we get:

```
country_name:   US
region_name:    CA
city_name:      Mountain View
```

Here we finally finish querying data.

---

LICENSE: [CC BY-NC 4.0](https://creativecommons.org/licenses/by-nc/4.0/)

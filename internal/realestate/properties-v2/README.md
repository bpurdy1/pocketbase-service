able name
COLUMNS
Add column
address
TEXT
NOT NULL
DEFAULT''
bathrooms
NUMERIC
NOT NULL
DEFAULT0
bedrooms
NUMERIC
NOT NULL
DEFAULT0
building_sf
NUMERIC
NOT NULL
DEFAULT0
city
TEXT
NOT NULL
DEFAULT''
county
TEXT
NOT NULL
DEFAULT''
created
TEXT
NOT NULL
DEFAULT''
id
TEXT
PRIMARY KEY
NOT NULL
DEFAULT'r'||lower(hex(randomblob(7)))
lat
NUMERIC
NOT NULL
DEFAULT0
lng
NUMERIC
NOT NULL
DEFAULT0
lot_sf
NUMERIC
NOT NULL
DEFAULT0
notes
TEXT
NOT NULL
DEFAULT''
number_of_units
NUMERIC
NOT NULL
DEFAULT0
organization
TEXT
NOT NULL
DEFAULT''
price
NUMERIC
NOT NULL
DEFAULT0
property_name
TEXT
NOT NULL
DEFAULT''
property_type
TEXT
NOT NULL
DEFAULT''
sqft
NUMERIC
NOT NULL
DEFAULT0
state
TEXT
NOT NULL
DEFAULT''

updated
TEXT
NOT NULL
DEFAULT''
year_built
NUMERIC
NOT NULL
DEFAULT0

zip_code
TEXT
NOT NULL
DEFAULT''
CONSTRAINTS
Add constraint
INDEXES
Add index
INDEX
idx_properties_created
…
(created)
INDEX
idx_properties_org
…
(organization)
AUTO-INDEXES

i need  sqlite realestate database which has 
this should be contain all infomation related to the property links to photos listing to where it was found. 

properties_listsing -> holds active  listing found 
-> unique per address
-> ttl 
-> quick update of ttl 
created at, modified at should be automatic 

properties_listsing_history -> historical of listsings at a aspecifc address 

rentail_listings > holds active  listing found 
rentail_listings_history -> historical of listsings at a aspecifc address 


search can be done via this 

I need to be able to search basesd on 
In SQLite, the R*Tr
ee module provides a specialized spatial index for performing fast range queries on multi-dimensional data, most commonly 2D geographical coordinates. 


please give me all tables.sqlite files and sqli

should this be done via sqlc?? to standardie everything
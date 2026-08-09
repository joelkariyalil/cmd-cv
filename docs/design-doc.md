Design Doc

Tech Stack: API Gateway + Lambdas + DynamoDB

Lambdas and their Responsibilities

GET - create-message-ui-lambda -> To Create the Message
POST - retrieve the unique ID by storing the message in DynamoDB with the TTL

retrieve-message-ui-lamdba -> To Retrieve the Message 
GET - retrieve-message-ui-lambda -> The webpage to basically provide the user to retrieve the message from the token or whatever
GET - the remaining TTL
DELETE - delete button to delete the message

Implementation Details


Unique ID - One Time Code (OTC): Pattern: [0-9]{6}


DynamoDB Data Model

PK, SK = UniqueID,TimeStamp

While the application itself, does not require the SK to be preset, I would like to add the SK in case, I would like to know the information such as, the total count or, how many OTCs were generated in some time etc...
If I do not have the SK then I would miss out on this opportunity.

I might drop this at a later stage, to not design for an access pattern not yet discovered :)



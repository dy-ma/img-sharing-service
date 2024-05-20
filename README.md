# Image Sharing Service

![Serverless Architecture Diagram](assets/image_sharing_service.png "Serverless Architecture Diagram")

## Supported Actions

- **[Upload]** Photographers create an Event in the UI. They can then upload 
images to the event by dragging and dropping from their local drive, or 
selecting with the iOS/Android photo selector. 
The photos are then bound to a randomly generated access code, which is given
to the photo subjects to view their photos at the specific event url.
- **[View]** A viewing user visits the link in their browser, and a text box 
appears, prompting for the access code.
Once authenticated, the photos in the album appear in a grid. 
The user has the option to download all, one, or a subset of the photos. 
There is a timer displaying the days/hours left before the photos are deleted.

## Data Model

Photo albums are oriented around "events", which contain a large set of photos.
Events have unique generated urls, and the album owners can share subsets of the
photos to specific people by generating "access codes".

Photographers can create an event, upload some images to it, and give subjects
the access code which lets them see only their photos.

Only photographers need to create an account to manage their events and upload
photos. Viewers of the photos only need to type in the access code.

### Database

### S3

## API

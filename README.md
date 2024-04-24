# Bubble

Full-stack application to create collaborative sessions on Spotify. Enables users to create listening parties and text invitations to their friends.
Anyone can then add songs to the host's queue.

## Development
- The backend is powered using Pocketbase, with custom Go extensions. Pocketbase was mainly used for its admin data interface.
- Users are signed into Spotify using OAuth 2.
- Front end is created using React Native and Expo
- Local backend is exposed to React Native frontend using ngrok.

GameFabric runs its own internal Container registry proxy, which is where you should push your game server images to in order to have them available in game servers.
Those images are scoped by branch. For example, a standard use case would be to have a development branch and a production branch.
The development branch would contain dev images to be used by a development Armada,
while the production branch would only contain releases of the game server that make it to production.

For details check the <a href="https://docs.gamefabric.com/multiplayer-servers/getting-started/glossary#branch">GameFabric documentation</a>.
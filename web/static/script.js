import forceGraph from 'https://cdn.skypack.dev/force-graph';

/**
 * @param {string} url
 * @returns {Promise<T>}
 */
function jsonFetch(url) {
  return fetch(url).then(resp => resp.json())
}

async function bootstrap() {

  const containers = await jsonFetch("/api/containers")
    .then(data => Promise.all(data.Containers.map(id => jsonFetch(`/api/containers/${id}`))));

  console.log("containers", containers);

  const networkIds = new Set()
  for (const { Networks } of containers) {
    Networks.forEach(id => networkIds.add(id));
  }

  const networks = await Promise.all(
    Array.from(networkIds)
      .map(id => jsonFetch(`/api/networks/${id}`))
  )

  console.log("networks", containers);

  const nodes = [];
  const links = [];

  for (const { ID, Name, Networks } of containers) {
    nodes.push({ id: ID, name: Name, type: "container" });
    for (const networkID of Networks) {
      links.push({ source: ID, target: networkID });
    }
  }

  for (const { ID, Name } of networks) {
    nodes.push({ id: ID, name: Name, type: "network" });
  }

  console.log("nodes", nodes);
  console.log("links", links);

  forceGraph()
    (document.getElementById('graph'))
    .graphData({ nodes, links });

}

await bootstrap();

console.log("yaay");

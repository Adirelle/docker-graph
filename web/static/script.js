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

  const networkIds = new Set();
  const volumeIds = new Set();

  const nodes = [];
  const links = [];

  for (const { ID, Name, Networks, Mounts } of containers) {
    nodes.push({ id: ID, name: Name, type: "container" });
    for (const networkID of Networks) {
      links.push({ source: ID, target: networkID });
      networkIds.add(networkID);
    }
    if (Mounts) {
      for (const { Type, Name, Source, Destination } of Mounts) {
        if (Type == "bind") {
          nodes.push({ id: Source, name: Destination, type: "bind" })
          links.push({ source: ID, target: Source });
        } else if (Type == "volume") {
          volumeIds.add(Name);
          links.push({ source: ID, target: Name });
        }
      }
    }
  }

  const networks = await Promise.all(
    Array.from(networkIds)
      .map(id => jsonFetch(`/api/networks/${id}`))
  )

  console.log("networks", networks);

  for (const { ID, Name } of networks) {
    nodes.push({ id: ID, name: Name, type: "network" });
  }

  const volumes = await Promise.all(
    Array.from(volumeIds)
      .map(id => jsonFetch(`/api/volumes/${id}`))
  )

  console.log("volumes", volumes);

  for (const { ID, Name } of volumes) {
    nodes.push({ id: ID, name: Name, type: "volume" });
  }

  console.log("nodes", nodes);
  console.log("links", links);

  forceGraph()
    (document.getElementById('graph'))
    .graphData({ nodes, links })
    .linkDirectionalArrowLength(6)
    .nodeCanvasObject(nodePaint);

}

/**
 *
 * @param {*} param0
 * @param {CanvasRenderingContext2D} ctx
 */
function nodePaint({ type, x, y }, ctx) {
  //ctx.fillStyle = color;
  switch (type) {
    case "network":
      ctx.beginPath();
      ctx.moveTo(x, y - 5);
      ctx.lineTo(x - 5, y + 5);
      ctx.lineTo(x + 5, y + 5);
      ctx.stroke();
      break;
    case "container":
      ctx.rect(x - 6, y - 4, 12, 8);
      ctx.stroke();
      break;
    default:
      ctx.beginPath();
      ctx.arc(x, y, 5, 0, 2 * Math.PI, false);
      ctx.stroke();
  }
}

await bootstrap();

console.log("yaay");

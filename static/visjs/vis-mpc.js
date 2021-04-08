fetch('/rpc')
    .then(response => response.json())
    .then(state => {
        //////////////NODES//////////////

        var node_off = (x_pos,y_pos,id_num, title_hover) => { return { id: id_num, shape: 'icon', icon: { face: "'FontAwesome'", code: "\uf1b2", size: 40, color: "black", }, borderWidth: 2, x: x_pos, y: y_pos, title:  title_hover}}
        var node_on = (x_pos,y_pos,id_num, title_hover, label_str) => { return { id: id_num, label: label_str, shape: 'image', image: "../cube.png", borderWidth: 2, x: x_pos, y: y_pos, title:  title_hover}}

        var nodes = new vis.DataSet([]);

        ///////////////EDGES/////////////

        var between_nodes = (from_node,to_node) => {return { from: from_node, to: to_node, color: "rgb(20,24,200)", arrows: "from, to" }}

        // create an array with edges
        var edges = new vis.DataSet([]);

        var updateNodesAndEdges = (edges, nodes,state, numNodesInserted) => {
            //State machine
            numNodes = state["Nodes"].length
            titles = []
            for(let i = 0; i < numNodes; i++){
                titles.push(buildTitle(state["Nodes"][i]["ID"], state["Nodes"][i]["Address"], state["Nodes"][i]["PendingJobs"], state["Nodes"][i]["DoneJobs"]))
            }
            for(let i = numNodesInserted; i < numNodes; i++){
                nodes.add(node_on(0, 0,i, titles[i], state["Nodes"][i]["Name"]));
                if (i > 0) {
                    for(let j = 0; j < i; j++){
                        edges.add(between_nodes(j,i));
                    }
                }
            }
            for(let i = 0; i < numNodesInserted; i++){
                nodes.update({ id: i, title: titles[i], label: state["Nodes"][i]["Name"]});
            }

            return numNodes
        }

        var buildTitle = (ID, address, pending, done) => {
            var NodeInfo = "ID: " + ID + "\n" + "Address: " + address + "\n"
            var PendingJobs = "\nPending Jobs \n"
            if (pending.length > 0){
                for (let i = 0; i < pending.length; i++){
                    PendingJobs += pending[i]["ID"] + "->" + pending[i]["Type"] + "\n"
                }
            }
            var DoneJobs = "\nDone Jobs \n"
            if (done.length > 0){
                for (let i = 0; i < done.length; i++){
                    DoneJobs += done[i]["ID"] + "->" + done[i]["Type"] + "\n"
                }
            }
            return NodeInfo + PendingJobs + DoneJobs
        }

        //updateNodes(state);

        // create a network
        var container = document.getElementById("mynetwork");
        var data = {
            nodes: nodes,
            edges: edges,
        };
        var options = {
            nodes: {
                shape: 'dot'
            },
            edges: {
                smooth: false
            },
            physics: true,
            interaction: {
                dragNodes: false,// do not allow dragging nodes
                zoomView: false, // do not allow zooming
                dragView: false  // do not allow dragging
            }
        };
        var network = new vis.Network(container, data, options);
        var numNodesInserted = 0;

        function timeout() {
            setTimeout(function () {
                fetch('/rpc')
                    .then(response => response.json())
                    .then(state => {
                        numNodesInserted = updateNodesAndEdges(edges, nodes, state, numNodesInserted);
                        var data = {
                            nodes: nodes,
                            edges: edges,
                        };
                    })
                timeout();
            }, 2000);
        };

        timeout()
    });


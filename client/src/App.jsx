import React, { useState, useEffect } from "react";
import { ClickRaceClient } from "./grpc/click_grpc_web_pb";
import { Click, Empty } from "./grpc/click_pb";

const client = new ClickRaceClient("http://localhost:8081");

function App() {
  const [name, setName] = useState("Player1");
  const [score, setScore] = useState(0);
  const [leaderboard, setLeaderboard] = useState([]);

  useEffect(() => {
    const stream = client.getLeaderboard(new Empty());
    stream.on("data", (resp) => {
      setLeaderboard(resp.getScoresList().map(s => ({
        player: s.getPlayer(),
        score: s.getScore()
      })));
    });
  }, []);

  const click = () => {
    const request = new Click();
    request.setPlayer(name);
    request.setAmount(1);
    const stream = client.sendClicks();
    stream.write(request);
    stream.end();
  };

  return (
    <div>
      <h1>Click Race</h1>
      <input value={name} onChange={(e) => setName(e.target.value)} />
      <button onClick={click}>Click!</button>
      <h2>Your Score: {score}</h2>
      <h3>Leaderboard</h3>
      <ul>
        {leaderboard.map((s, i) => (
          <li key={i}>{s.player}: {s.score}</li>
        ))}
      </ul>
    </div>
  );
}

export default App;

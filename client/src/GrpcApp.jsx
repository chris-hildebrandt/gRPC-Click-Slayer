import React, { useState, useEffect, useRef } from 'react';
import './App.css';

// Import gRPC-Web generated clients/messages
const grpc = require('grpc-web');
const { MonsterSlayerClient } = require('./grpc/monster_slayer_grpc_web_pb');
const {
  JoinGameRequest,
  AttackRequest,
  GameUpdate,
} = require('./grpc/monster_slayer_pb');
const { Empty } = require('google-protobuf/google/protobuf/empty_pb.js');

function GrpcApp() {
  const [playerName, setPlayerName] = useState('');
  const [isJoined, setIsJoined] = useState(false);
  const [player, setPlayer] = useState(null);
  const [monster, setMonster] = useState(null);
  const [players, setPlayers] = useState([]);
  const [isConnected, setIsConnected] = useState(false);
  const [clickCount, setClickCount] = useState(0);
  
  const clientRef = useRef(null);
  const streamRef = useRef(null);

  // Initialize gRPC client
  useEffect(() => {
    const client = new MonsterSlayerClient('http://localhost:8081', null, null);
    clientRef.current = client;
  }, []);

  const joinGame = async () => {
    if (!playerName.trim()) return;

    try {
      const request = new JoinGameRequest();
      request.setPlayerName(playerName);

      clientRef.current.joinGame(request, {}, (err, resp) => {
        if (err) {
          console.error('JoinGame error:', err);
          alert('Failed to join game. Make sure Envoy and server are running.');
          return;
        }
        const playerObj = resp.getPlayer().toObject();
        const monsterObj = resp.getMonster().toObject();
        setPlayer(playerObj);
        setMonster(monsterObj);
        setIsJoined(true);
        setIsConnected(true);

        // start streaming updates
        const empty = new Empty();
        const stream = clientRef.current.streamGameUpdates(empty, {});
        streamRef.current = stream;
        stream.on('data', (update) => {
          const u = update.toObject();
          setMonster(u.monster);
          setPlayers(u.playersList || []);
        });
        stream.on('error', (e) => {
          console.error('Stream error:', e);
          setIsConnected(false);
        });
        stream.on('end', () => {
          setIsConnected(false);
        });
      });
    } catch (error) {
      console.error('Error joining game:', error);
      alert('Failed to join game. Make sure Envoy and server are running.');
    }
  };

  const attackMonster = async () => {
    if (!player || !monster) return;

    const damage = 10; // Base damage per click
    setClickCount(prev => prev + 1);

    try {
      const request = new AttackRequest();
      request.setPlayerName(player.name);
      request.setDamage(damage);

      clientRef.current.attackMonster(request, {}, (err, resp) => {
        if (err) {
          console.error('AttackMonster error:', err);
          return;
        }
        const data = resp.toObject();
        setPlayer(data.player);
        setMonster(data.monster);
        if (data.monsterDefeated) {
          alert('🎉 Monster defeated! You got loot: ' + data.lootGained);
        }
      });
    } catch (error) {
      console.error('Error attacking monster:', error);
    }
  };

  const getHealthPercentage = () => {
    if (!monster) return 0;
    return (monster.current_health / monster.max_health) * 100;
  };

  const getSortedPlayers = () => {
    return [...players].sort((a, b) => (b.totalClicks || 0) - (a.totalClicks || 0));
  };

  if (!isJoined) {
    return (
      <div className="app">
        <div className="join-screen">
          <h1>🐉 Monster Slayer (gRPC)</h1>
          <p>Join the battle against the mighty dragon using gRPC!</p>
          <div className="join-form">
            <input
              type="text"
              placeholder="Enter your name"
              value={playerName}
              onChange={(e) => setPlayerName(e.target.value)}
              onKeyPress={(e) => e.key === 'Enter' && joinGame()}
            />
            <button onClick={joinGame} disabled={!playerName.trim()}>
              Join Battle
            </button>
          </div>
          <div className="tech-info">
            <p><strong>Technologies:</strong> gRPC, Protocol Buffers, Go, React</p>
            <p><strong>Features:</strong> Real-time streaming, Type-safe communication</p>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="app">
      <header className="app-header">
        <h1>🐉 Monster Slayer (gRPC)</h1>
        <div className="connection-status">
          {isConnected ? '🟢 gRPC Connected' : '🔴 Disconnected'}
        </div>
      </header>

      <div className="game-container">
        {/* Monster Section */}
        <div className="monster-section">
          <h2>Current Monster: {monster?.name}</h2>
          <div className="monster-health">
            <div className="health-bar">
              <div 
                className="health-fill" 
                style={{ width: `${getHealthPercentage()}%` }}
              ></div>
            </div>
            <div className="health-text">
              {monster?.currentHealth} / {monster?.maxHealth} HP
            </div>
          </div>
          <div className="monster-level">Level {monster?.level}</div>
        </div>

        {/* Player Section */}
        <div className="player-section">
          <h2>Your Stats</h2>
          <div className="player-stats">
            <div>Name: {player?.name}</div>
            <div>Total Clicks: {player?.totalClicks}</div>
            <div>Level: {player?.level}</div>
            <div>Experience: {player?.experience}</div>
            <div>Loot: {(player?.lootCollected || []).join(', ') || 'None'}</div>
          </div>
          
          <button 
            className="click-button" 
            onClick={attackMonster}
            disabled={!isConnected}
          >
            ⚔️ Attack! ({clickCount})
          </button>
        </div>

        {/* Leaderboard */}
        <div className="leaderboard-section">
          <h2>Leaderboard</h2>
          <div className="leaderboard">
            {getSortedPlayers().map((p, index) => (
              <div key={p.name} className={`leaderboard-item ${p.name === player?.name ? 'current-player' : ''}`}>
                <span className="rank">#{index + 1}</span>
                <span className="name">{p.name}</span>
                <span className="clicks">{p.totalClicks} clicks</span>
                <span className="level">Lv.{p.level}</span>
              </div>
            ))}
          </div>
        </div>

        {/* gRPC Info */}
        <div className="grpc-info">
          <h3>gRPC Features Demonstrated</h3>
          <ul>
            <li>✅ <strong>Unary RPC:</strong> JoinGame, AttackMonster</li>
            <li>✅ <strong>Server Streaming:</strong> StreamGameUpdates</li>
            <li>✅ <strong>Protocol Buffers:</strong> Type-safe messages</li>
            <li>✅ <strong>Go Backend:</strong> Concurrent server</li>
            <li>✅ <strong>React Frontend:</strong> Real-time UI</li>
          </ul>
        </div>
      </div>
    </div>
  );
}

export default GrpcApp;

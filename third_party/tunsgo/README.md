<div align="center">
  <img src="https://img.shields.io/badge/TunsGo-P2P_Proxy-blueviolet?style=for-the-badge&logo=go" alt="TunsGo Logo" />
  
  <h1>🌐 TunsGo</h1>
  <p><b>High-performance, decentralized P2P proxy routing engine built on libp2p</b></p>

  <p>
    <a href="https://github.com/YouROK/tunsgo/releases">
      <img src="https://img.shields.io/github/v/tag/YouROK/tunsgo?label=version&color=blue&style=flat-square" alt="version">
    </a>
    <img src="https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat-square&logo=go" alt="Go Version">
    <img src="https://img.shields.io/badge/License-GPLv3-blue.svg?style=flat-square" alt="GPLv3 License">
  </p>
</div>

<hr />

<h2>📖 Description</h2>
<p>
  TunsGo is a decentralized mesh-network proxy. It allows nodes to discover each other and share access to specific hosts, 
  routing traffic through a distributed network without any central authority. Every node acts as both a consumer and a provider.
</p>

<h2>🚀 Installation</h2>
<p>Install the binary directly to your system:</p>
<pre><code>go install github.com/YouROK/tunsgo/cmd/tuns@latest</code></pre>

<h2>🛠 Quick Start</h2>
<ol>
  <li><b>Launch:</b> Run <code>tuns</code> in your terminal.</li>
  <li><b>Configure:</b> On first run, it generates <code>tuns.conf</code> in the current directory.</li>
  <li><b>Proxy:</b> The server starts on <code>:8080</code> by default.</li>
</ol>

<hr />

<h2>⚙️ Configuration</h2>
<details open>
  <summary><b>tuns.conf (YAML)</b></summary>

```yaml
server:
  port: "8080"        # Local HTTP gateway
  slots: 5            # Concurrent request workers
  slot_sleep: 1       # Throttle delay (seconds)

p2p:
  low_conns: 20       # Minimum neighbors to maintain
  hi_conns: 50        # Connection cap

provided_hosts:       # Domains you share with the network
  - "*themoviedb.org"
  - "*tmdb.org"
```
</details>
<hr />

<h2>📡 API Reference</h2>
<table width="100%">
  <thead>
    <tr>
      <th align="left">Endpoint</th>
      <th align="left">Method</th>
      <th align="left">Description</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td><code>/proxy/:url</code></td>
      <td><code>ANY</code></td>
      <td>Routes a full URL through the P2P network</td>
    </tr>
    <tr>
      <td><code>/status</code></td>
      <td><code>GET</code></td>
      <td>Returns node health and peer statistics</td>
    </tr>
  </tbody>
</table>

<p><b>Example Proxy Call:</b></p>
<pre><code>curl http://localhost:8080/proxy/https://api.themoviedb.org/3/movie/550</code></pre>

<hr />

<div align="center">
  <p>
    <img src="https://img.shields.io/badge/License-GPLv3-blue.svg?style=flat-square" alt="GPLv3 License">
  </p>
  <sub>
    Released under the <b>GNU General Public License v3.0</b>.<br />
    Built with ❤️ using <a href="https://github.com/libp2p/libp2p">libp2p</a>.
  </sub>
</div>

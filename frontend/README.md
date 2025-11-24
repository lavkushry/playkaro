# PlayKaro Frontend

React 19 + Vite + Tailwind UI for the PlayKaro sportsbook, casino launcher, wallet, and admin tools.

## Prerequisites
- Node.js 18+
- Backend API reachable at `http://localhost:8080` (default). The API base URLs are hard-coded; update the files listed below if you host the backend elsewhere.

## Run
```bash
npm install
npm run dev      # start on :5173
npm run build    # production build
npm run preview  # serve the built app locally
npm run lint     # eslint
```

## API Endpoints Used by the UI
- REST base: `http://localhost:8080/api/v1` in `src/store/useAuthStore.js`, `src/store/useWalletStore.js`, `src/store/useBetStore.js`.
- GraphQL: `http://localhost:8080/query` in `src/apolloClient.js`.
- WebSocket: `ws://localhost:8080/ws` inside `useBetStore` for live odds and chat.
Update those constants if you change the backend host/port.

## Screens & Features
- **Auth**: Login/Register forms with JWT storage in `localStorage`.
- **Dashboard**: Balance widget, promotion carousel, quick actions.
- **Sportsbook**: Live matches, bet slip with amount/odds, WebSocket odds updates.
- **Admin**: Manage matches/odds/settlement (requires admin flag in DB).
- **Wallet**: Deposit/withdraw UI with optimistic updates.
- **KYC**: Upload documents and view status.
- **Casino**: Catalog of seeded games and launch links.
- **Promotions/Leaderboard/Analytics**: Bonus claims, referrals, and leaderboard views.

## Code Pointers
- State stores: `src/store` (auth, wallet, bets).
- GraphQL operations: `src/graphql/queries.js`.
- Global styles/theme: `src/index.css`, `src/App.css`.
- Shared UI components: `src/components/ui`.

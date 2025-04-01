import { Keypair } from "@solana/web3.js";

const keypair = Keypair.generate();
console.log("address:", keypair.publicKey.toBase58());
console.log("keypair:", keypair.secretKey.toString());
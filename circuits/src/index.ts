import { Circomkit } from "circomkit";

async function main() {
    // create circomkit
    const circomkit = new Circomkit({
        protocol: "groth16",
    });

    const proof = await circomkit.prove("singleTx", "test");

    console.log("Proof generated:", proof);

    const ok = await circomkit.verify("singleTx","test");

    if (ok) {
        console.log("Proof is valid", "success");
    } else {
        console.log("Proof is invalid", "error");
    }
}

main()
  .then(() => process.exit(0))
  .catch((e) => {
    console.error(e);
    process.exit(1);
  });
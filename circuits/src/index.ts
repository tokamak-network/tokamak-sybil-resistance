import { Circomkit } from "circomkit";

async function main() {
    // create circomkit
    const circomkit = new Circomkit({
        protocol: "groth16",
    });

    await circomkit.compile("src/singleTx", {
        file:"singleTx",
        template:"SingleTx",
        params:[1],
    });

    await circomkit.prove("singleTx", "singleTx_input");

    const ok = await circomkit.verify("singleTx","singleTx_input");

    if (ok) {
        console.log("Proof is valid", "success");
    } else {
        console.log("Proof is invalid", "error");
    }
}
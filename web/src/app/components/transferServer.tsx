import {Note} from "@/app/components/formSubcomponents/notes";
import {TestNotes, ExampleImageLocation} from "@/app/components/formSubcomponents/contaminations";
import {ACL} from "@/app/components/accessControlServer";


export function TestTransferOk(){
    return new TransferData({
        _id: "(TRANSFER ID HERE)",
        from: "(FROM 1)",
        to: "(TO)",
        fromType: "plate",
        toType: "plate",
        creationDate: Date.now()-10000,
        reason: "mold",
        fromImage: ExampleImageLocation,
        toImage: ExampleImageLocation,
        notes: TestNotes,
        lastUpdated: 789,
        //perms: {userPerms: {ids:[{id:"userCollId",val:"userName"}],canWrite:[true]},projectPerms: {ids:["proj1","proj2"],canWrite:[true, false]}, blanketPerms: 1},
    })
}

export interface TransferData {
    _id: string
    from: string
    to: string
    fromType: string
    toType: string
    creationDate: number
    reason: string
    fromImage?: string
    toImage?: string
    notes?: Note[]
    lastUpdated: number
    acl?: ACL
}
export class TransferData {
    // Accept a single object containing the fields
    constructor(init?: Partial<TransferData>) {
        // Dynamically map the object fields onto the class instance
        Object.assign(this, init);
    }

    public getId(): string {
        return this._id
    }
    public entryType(): string {
        return "transfer"
    }
}


// Likely will never need a closeable single selector for this
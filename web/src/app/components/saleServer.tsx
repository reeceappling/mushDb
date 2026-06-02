import {Note} from "@/app/components/formSubcomponents/notes";
import {ACL} from "@/app/components/accessControlServer";

export function TestSaleOk() {
    return new SaleData({
        _id: "(SALE ID HERE)",
        creationDate: 123,
        notes: [{
            time: 123,
            note: "(NOTE 1)"
        }, {
            time: 456,
            note: "(NOTE 2)"
        }],
        lastUpdated: 789,
        //perms: {userPerms: {ids:[{id:"userCollId",val:"userName"}],canWrite:[true]},projectPerms: {ids:["proj1","proj2"],canWrite:[true, false]}, blanketPerms: 1},
    })
}

export interface ItemWithNumber {
    _id: string // item id
    n: number
}

export interface SaleData {
    _id: string // lot number
    creationDate: number
    //itemsSold: ItemWithNumber[] // TODO: consider doing this!
    notes?: Note[]
    lastUpdated: number
    acl?: ACL
}
export class SaleData {
    // Accept a single object containing the fields
    constructor(init?: Partial<SaleData>) {
        // Dynamically map the object fields onto the class instance
        Object.assign(this, init);
    }

    public getId(): string {
        return this._id
    }
    public entryType(): string {
        return "sale"
    }
}

// // TODO: NECESSARY?
// export function SaleSelector(sp: SelectorProps<SaleData>){
//     // TODO: REDO?
//     // return RecentSelector<SaleData>({
//     //     recentEndpt: "sales",
//     //     assertType: AssertSale,
//     //     closeTxt: "Close Sale List",
//     //     createTxt: "Create Sale",
//     //     newForm: NewSaleForm,
//     //     createEndpt: "sale",
//     //     lowercase: "sale",
//     //     inline: (inlineIn: InlineProps<SaleData>)=>{return SaleInline(inlineIn)},
//     // })(sp)
// }

// export const ChannelTextNewSale = "newSale"

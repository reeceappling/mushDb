import {Note} from "@/app/components/formSubcomponents/notes";
import {ACL, TestAcl} from "@/app/components/accessControlServer";

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
        acl: TestAcl(),
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
    acl: ACL
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
    public description(): string {
        return `Sale ${this._id}. Created on ${new Date(this.creationDate).toISOString()}. Last updated on ${new Date(this.lastUpdated).toISOString()}.` // TODO: KF, FilterSize, flushes, contams, disposal date, etc?
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

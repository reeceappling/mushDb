import {Note} from "@/app/components/formSubcomponents/notes";
import {AssertSale, NewSaleForm, SaleInline} from "@/app/components/saleClient";
import RecentSelector, {SelectorProps} from "@/app/components/selector";
import {InlineProps} from "@/app/components/common";
import {EntryPerms} from "@/app/components/perms";
import {ACL} from "@/app/components/accessControlServer";

export function TestSaleOk() {
    const a: SaleData = {
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
    }
    return a
}

export interface SaleData {
    _id: string // lot number
    creationDate: number
    notes?: Note[]
    lastUpdated: number
    acl?: ACL
}

// // TODO: NECESSARY?
// export function SaleSelector(sp: SelectorProps<SaleData>){
//     // TODO: REDO?
//     // return RecentSelector<SaleData>({
//     //     msgTxt: ChannelTextNewSale,
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

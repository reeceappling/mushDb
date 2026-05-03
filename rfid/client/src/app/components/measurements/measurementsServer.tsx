export type AgarColor = "clear" | "black" | "blue" | "green" | "red";

// export interface AgarBatchData {
//     _id: string
//     color: string
//     pcRun: string
//     recipe: string
//     notes?: Note[]
//     lastUpdated: number
// }
//
// export function AgarBatchSelector(sp: SelectorProps<AgarBatchData>) {
//     // TODO: LOOK UP MOST RECENT AGAR BATCHES?
//     // TODO: SELECT FROM THOSE BATCHES?
//     // TODO: DONT USE RECENTSELECTOR!?
//     return <RecentSelector props={{
//         allowCreation: sp.allowCreation,
//         doSelect: sp.doSelect, // For selecting normally
//         msgTxt: ChannelTextNewAgarBatch,
//         recentEndpt: "agarBatches",
//         assertType: AssertAgarBatch,
//         closeTxt: "Close Batch List",
//         createTxt: "Create Agar Batch",
//         createEndpt: "agarBatch",
//         lowercase: "agar batch",
//         creatorInPage: sp.creatorInPage,
//         inline: (inn: InlineProps<AgarBatchData>) => {
//             return <AgarBatchInline data={inn.data} headerLevel={inn.headerLevel} onClick={inn.onClick}
//                                     expandByDefault={inn.expandByDefault}/>
//         },
//         getId: (v: AgarBatchData) => {
//             return v._id
//         }
//     }}>
//         <NewAgarBatchForm redirectOnCreate={false} onCreate={sp.doSelect}/>
//     </RecentSelector>
// }
//
// export const ChannelTextNewAgarBatch = "newAgarBatch"